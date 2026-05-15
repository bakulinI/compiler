package parser

import (
	"fmt"

	"compiler_labs/internal/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
	errors []SyntaxError
}

type ParseResult struct {
	AST            *ASTNode
	ErrorMessages  []string
	SuccessMessage string
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
		errors: make([]SyntaxError, 0),
	}
}

func (p *Parser) Parse() *ParseResult {
	ast := p.parseProgram()

	if !p.isAtEnd() {
		token := p.current()
		p.addError("ЛИШНИЙ ТОКЕН", "Лишний токен после конца программы", "", token)
	}

	result := &ParseResult{
		AST:           ast,
		ErrorMessages: make([]string, 0),
	}

	for _, err := range p.errors {
		result.ErrorMessages = append(result.ErrorMessages, err.String())
	}

	if len(result.ErrorMessages) == 0 {
		result.SuccessMessage = "Синтаксический анализ завершён успешно. Ошибок не найдено."
	} else {
		result.SuccessMessage = fmt.Sprintf("Синтаксический анализ завершён с ошибками. Найдено ошибок: %d.", len(result.ErrorMessages))
	}

	return result
}

func (p *Parser) parseProgram() *ASTNode {
	node := NewNode("Program")

	p.consume(lexer.KEYWORD, "package", "ключевое слово package")
	if name, ok := p.consume(lexer.IDENTIFIER, "", "имя пакета"); ok {
		node.AddField("name", name.Value)
	}

	imports := NewNode("Imports")
	for p.check(lexer.KEYWORD, "import") {
		imports.AddChild(p.parseImportDecl())
	}
	if len(imports.Children) > 0 {
		node.AddChild(imports)
	}

	declarations := NewNode("Declarations")
	for !p.isAtEnd() {
		if p.check(lexer.KEYWORD, "func") {
			declarations.AddChild(p.parseFuncDecl())
			continue
		}

		token := p.current()
		p.addError("НЕОЖИДАННЫЙ ТОКЕН", "Ожидалось объявление функции верхнего уровня", "func", token)
		p.advance()
	}
	node.AddChild(declarations)

	return node
}

func (p *Parser) parseImportDecl() *ASTNode {
	node := NewNode("ImportDecl")
	p.consume(lexer.KEYWORD, "import", "ключевое слово import")
	if path, ok := p.consume(lexer.CONSTANT_STR, "", "строковый путь импортируемого пакета"); ok {
		node.AddField("path", path.Value)
	}
	return node
}

func (p *Parser) parseFuncDecl() *ASTNode {
	node := NewNode("FuncDecl")
	p.consume(lexer.KEYWORD, "func", "ключевое слово func")

	if name, ok := p.consume(lexer.IDENTIFIER, "", "имя функции"); ok {
		node.AddField("name", name.Value)
	}

	node.AddChild(p.parseParameterList())

	if p.isTypeStart() {
		returnType := p.advance()
		node.AddField("return_type", returnType.Value)
	}

	node.AddChild(p.parseBlock())
	return node
}

func (p *Parser) parseParameterList() *ASTNode {
	node := NewNode("Parameters")
	p.consume(lexer.DELIMITER, "(", "открывающая скобка параметров")

	if !p.check(lexer.DELIMITER, ")") && !p.isAtEnd() {
		for {
			param := NewNode("Param")
			if name, ok := p.consume(lexer.IDENTIFIER, "", "имя параметра"); ok {
				param.AddField("name", name.Value)
			}

			if typ, ok := p.consumeType("тип параметра"); ok {
				param.AddField("type", typ.Value)
			}

			node.AddChild(param)
			if !p.match(lexer.DELIMITER, ",") {
				break
			}
		}
	}

	p.consume(lexer.DELIMITER, ")", "закрывающая скобка параметров")
	return node
}

func (p *Parser) parseBlock() *ASTNode {
	node := NewNode("Block")
	if _, ok := p.consume(lexer.DELIMITER, "{", "открывающая фигурная скобка блока"); !ok {
		return node
	}

	for !p.isAtEnd() && !p.check(lexer.DELIMITER, "}") {
		if p.match(lexer.DELIMITER, ";") {
			continue
		}
		node.AddChild(p.parseStatement())
	}

	p.consume(lexer.DELIMITER, "}", "закрывающая фигурная скобка блока")
	return node
}

func (p *Parser) parseStatement() *ASTNode {
	if p.check(lexer.KEYWORD, "var") {
		return p.parseVarDecl()
	}
	if p.check(lexer.KEYWORD, "return") {
		return p.parseReturnStmt()
	}
	if p.check(lexer.KEYWORD, "if") {
		return p.parseIfStmt()
	}
	if p.check(lexer.DELIMITER, "{") {
		return p.parseBlock()
	}
	if p.isAssignmentStart() {
		return p.parseAssignStmt()
	}

	return p.parseExprStmt()
}

func (p *Parser) parseVarDecl() *ASTNode {
	node := NewNode("VarDecl")
	p.consume(lexer.KEYWORD, "var", "ключевое слово var")

	if name, ok := p.consume(lexer.IDENTIFIER, "", "имя переменной"); ok {
		node.AddField("name", name.Value)
	}
	if typ, ok := p.consumeType("тип переменной"); ok {
		node.AddField("type", typ.Value)
	}
	if p.match(lexer.OPERATOR, "=") {
		node.AddChild(namedChild("value", p.parseExpression()))
	}

	return node
}

func (p *Parser) parseAssignStmt() *ASTNode {
	node := NewNode("AssignStmt")
	if name, ok := p.consume(lexer.IDENTIFIER, "", "левая часть присваивания"); ok {
		node.AddField("left", name.Value)
	}
	if op, ok := p.consumeAnyOperator([]string{"=", ":="}, "оператор присваивания"); ok {
		node.AddField("operator", op.Value)
	}
	node.AddChild(namedChild("right", p.parseExpression()))
	return node
}

func (p *Parser) parseReturnStmt() *ASTNode {
	node := NewNode("ReturnStmt")
	p.consume(lexer.KEYWORD, "return", "ключевое слово return")

	if !p.isAtEnd() && !p.check(lexer.DELIMITER, "}") && !p.check(lexer.DELIMITER, ";") {
		node.AddChild(namedChild("value", p.parseExpression()))
	}
	return node
}

func (p *Parser) parseIfStmt() *ASTNode {
	node := NewNode("IfStmt")
	p.consume(lexer.KEYWORD, "if", "ключевое слово if")
	node.AddChild(namedChild("condition", p.parseExpression()))
	node.AddChild(namedChild("then", p.parseBlock()))

	if p.match(lexer.KEYWORD, "else") {
		node.AddChild(namedChild("else", p.parseBlock()))
	}
	return node
}

func (p *Parser) parseExprStmt() *ASTNode {
	node := NewNode("ExprStmt")
	node.AddChild(p.parseExpression())
	return node
}

func (p *Parser) parseExpression() *ASTNode {
	return p.parseLogicalOr()
}

func (p *Parser) parseLogicalOr() *ASTNode {
	node := p.parseLogicalAnd()
	for p.match(lexer.OPERATOR, "||") {
		node = p.binaryNode("||", node, p.parseLogicalAnd())
	}
	return node
}

func (p *Parser) parseLogicalAnd() *ASTNode {
	node := p.parseEquality()
	for p.match(lexer.OPERATOR, "&&") {
		node = p.binaryNode("&&", node, p.parseEquality())
	}
	return node
}

func (p *Parser) parseEquality() *ASTNode {
	node := p.parseComparison()
	for p.checkAnyOperator([]string{"==", "!="}) {
		op := p.advance()
		node = p.binaryNode(op.Value, node, p.parseComparison())
	}
	return node
}

func (p *Parser) parseComparison() *ASTNode {
	node := p.parseTerm()
	for p.checkAnyOperator([]string{"<", "<=", ">", ">="}) {
		op := p.advance()
		node = p.binaryNode(op.Value, node, p.parseTerm())
	}
	return node
}

func (p *Parser) parseTerm() *ASTNode {
	node := p.parseFactor()
	for p.checkAnyOperator([]string{"+", "-"}) {
		op := p.advance()
		node = p.binaryNode(op.Value, node, p.parseFactor())
	}
	return node
}

func (p *Parser) parseFactor() *ASTNode {
	node := p.parseUnary()
	for p.checkAnyOperator([]string{"*", "/", "%"}) {
		op := p.advance()
		node = p.binaryNode(op.Value, node, p.parseUnary())
	}
	return node
}

func (p *Parser) parseUnary() *ASTNode {
	if p.checkAnyOperator([]string{"!", "-"}) {
		op := p.advance()
		node := NewNode("UnaryExpr")
		node.AddField("operator", op.Value)
		node.AddChild(p.parseUnary())
		return node
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() *ASTNode {
	node := p.parsePrimary()

	for {
		if p.match(lexer.DELIMITER, ".") {
			selector := NewNode("SelectorExpr")
			selector.AddChild(namedChild("object", node))
			if field, ok := p.consume(lexer.IDENTIFIER, "", "имя поля или метода после точки"); ok {
				selector.AddField("field", field.Value)
			}
			node = selector
			continue
		}

		if p.match(lexer.DELIMITER, "(") {
			call := NewNode("CallExpr")
			call.AddChild(namedChild("callee", node))

			args := NewNode("Arguments")
			if !p.check(lexer.DELIMITER, ")") && !p.isAtEnd() {
				for {
					args.AddChild(p.parseExpression())
					if !p.match(lexer.DELIMITER, ",") {
						break
					}
				}
			}
			call.AddChild(args)
			p.consume(lexer.DELIMITER, ")", "закрывающая скобка вызова")
			node = call
			continue
		}

		break
	}

	return node
}

func (p *Parser) parsePrimary() *ASTNode {
	if p.isAtEnd() {
		token := p.previous()
		p.addError("НЕОЖИДАННЫЙ КОНЕЦ", "Ожидалось выражение", "выражение", token)
		return NewNode("MissingExpr")
	}

	token := p.current()
	switch token.Type {
	case lexer.IDENTIFIER:
		p.advance()
		node := NewNode("Identifier")
		node.AddField("name", token.Value)
		return node
	case lexer.CONSTANT_INT:
		p.advance()
		node := NewNode("IntLiteral")
		node.AddField("value", token.Value)
		return node
	case lexer.CONSTANT_REAL:
		p.advance()
		node := NewNode("RealLiteral")
		node.AddField("value", token.Value)
		return node
	case lexer.CONSTANT_STR:
		p.advance()
		node := NewNode("StringLiteral")
		node.AddField("value", token.Value)
		return node
	case lexer.CONSTANT_BOOL:
		p.advance()
		node := NewNode("BoolLiteral")
		node.AddField("value", token.Value)
		return node
	}

	if p.match(lexer.DELIMITER, "(") {
		node := NewNode("GroupedExpr")
		node.AddChild(p.parseExpression())
		p.consume(lexer.DELIMITER, ")", "закрывающая скобка выражения")
		return node
	}

	p.addError("НЕОЖИДАННЫЙ ТОКЕН", "Ожидалось выражение", "идентификатор, константа или '('", token)
	p.advance()
	return NewNode("InvalidExpr")
}

func (p *Parser) binaryNode(operator string, left, right *ASTNode) *ASTNode {
	node := NewNode("BinaryExpr")
	node.AddField("operator", operator)
	node.AddChild(namedChild("left", left))
	node.AddChild(namedChild("right", right))
	return node
}

func namedChild(name string, child *ASTNode) *ASTNode {
	node := NewNode(name)
	node.AddChild(child)
	return node
}

func (p *Parser) consumeType(expected string) (lexer.Token, bool) {
	if p.isTypeStart() {
		return p.advance(), true
	}
	p.addError("ОТСУТСТВУЕТ ТИП", "Ожидался тип", expected, p.current())
	return lexer.Token{}, false
}

func (p *Parser) consume(tokenType lexer.TokenType, value, expected string) (lexer.Token, bool) {
	if p.check(tokenType, value) {
		return p.advance(), true
	}

	p.addError("НЕОЖИДАННЫЙ ТОКЕН", "Нарушена ожидаемая структура программы", expected, p.current())
	return lexer.Token{}, false
}

func (p *Parser) consumeAnyOperator(values []string, expected string) (lexer.Token, bool) {
	if p.checkAnyOperator(values) {
		return p.advance(), true
	}
	p.addError("НЕОЖИДАННЫЙ ОПЕРАТОР", "Ожидался один из допустимых операторов", expected, p.current())
	return lexer.Token{}, false
}

func (p *Parser) match(tokenType lexer.TokenType, value string) bool {
	if !p.check(tokenType, value) {
		return false
	}
	p.advance()
	return true
}

func (p *Parser) check(tokenType lexer.TokenType, value string) bool {
	if p.isAtEnd() {
		return false
	}

	token := p.current()
	if token.Type != tokenType {
		return false
	}
	return value == "" || token.Value == value
}

func (p *Parser) checkAnyOperator(values []string) bool {
	if p.isAtEnd() || p.current().Type != lexer.OPERATOR {
		return false
	}
	for _, value := range values {
		if p.current().Value == value {
			return true
		}
	}
	return false
}

func (p *Parser) isAssignmentStart() bool {
	return p.pos+1 < len(p.tokens) &&
		p.tokens[p.pos].Type == lexer.IDENTIFIER &&
		p.tokens[p.pos+1].Type == lexer.OPERATOR &&
		(p.tokens[p.pos+1].Value == "=" || p.tokens[p.pos+1].Value == ":=")
}

func (p *Parser) isTypeStart() bool {
	if p.isAtEnd() {
		return false
	}
	token := p.current()
	if token.Type != lexer.KEYWORD && token.Type != lexer.IDENTIFIER {
		return false
	}

	switch token.Value {
	case "int", "float64", "bool", "string":
		return true
	default:
		return token.Type == lexer.IDENTIFIER
	}
}

func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.pos++
	}
	return p.previous()
}

func (p *Parser) current() lexer.Token {
	if p.isAtEnd() {
		if len(p.tokens) == 0 {
			return lexer.Token{Line: 1, Column: 1, Value: "<EOF>"}
		}
		last := p.tokens[len(p.tokens)-1]
		return lexer.Token{Line: last.Line, Column: last.Column + len(last.Value), Value: "<EOF>"}
	}
	return p.tokens[p.pos]
}

func (p *Parser) previous() lexer.Token {
	if p.pos == 0 {
		return p.current()
	}
	return p.tokens[p.pos-1]
}

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens)
}

func (p *Parser) addError(errorType, message, expected string, token lexer.Token) {
	p.errors = append(p.errors, SyntaxError{
		Type:     errorType,
		Message:  message,
		Line:     token.Line,
		Column:   token.Column,
		Expected: expected,
		Actual:   fmt.Sprintf("%s '%s'", token.Type, token.Value),
	})
}
