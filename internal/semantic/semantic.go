package semantic

import (
	"fmt"
	"path"
	"strings"

	"compiler_labs/internal/parser"
)

type Analyzer struct {
	root    *parser.ASTNode
	scope   *scope
	symbols []*Symbol
	errors  []SemanticError
	triads  []Triad
}

type scope struct {
	name    string
	parent  *scope
	symbols map[string]*Symbol
}

type exprResult struct {
	typ   string
	place string
}

func NewAnalyzer(root *parser.ASTNode) *Analyzer {
	return &Analyzer{root: root}
}

func (a *Analyzer) Analyze() *Result {
	a.scope = &scope{name: "global", symbols: make(map[string]*Symbol)}
	a.visitProgram(a.root)

	result := &Result{
		Symbols:       make([]Symbol, 0, len(a.symbols)),
		Triads:        append([]Triad(nil), a.triads...),
		ErrorMessages: make([]string, 0, len(a.errors)),
	}

	for _, symbol := range a.symbols {
		result.Symbols = append(result.Symbols, *symbol)
	}
	for _, err := range a.errors {
		result.ErrorMessages = append(result.ErrorMessages, err.String())
	}

	if len(result.ErrorMessages) == 0 {
		result.SuccessMessage = "Семантический анализ завершён успешно. Ошибок не найдено."
	} else {
		result.SuccessMessage = fmt.Sprintf("Семантический анализ завершён с ошибками. Найдено ошибок: %d.", len(result.ErrorMessages))
	}
	return result
}

func (a *Analyzer) visitProgram(node *parser.ASTNode) {
	if node == nil {
		return
	}

	for _, imports := range node.Children {
		if imports.Name != "Imports" {
			continue
		}
		for _, importDecl := range imports.Children {
			a.visitImport(importDecl)
		}
	}

	for _, declarations := range node.Children {
		if declarations.Name != "Declarations" {
			continue
		}
		for _, declaration := range declarations.Children {
			if declaration.Name == "FuncDecl" {
				a.declare(functionSymbol(declaration), declaration)
			}
		}
		for _, declaration := range declarations.Children {
			if declaration.Name == "FuncDecl" {
				a.visitFuncDecl(declaration)
			}
		}
	}
}

func (a *Analyzer) visitImport(node *parser.ASTNode) {
	importPath := strings.Trim(fieldValue(node, "path"), "\"`'")
	if importPath == "" {
		return
	}
	name := path.Base(importPath)
	a.declare(&Symbol{
		Name:            name,
		Type:            "package",
		Kind:            SymbolPackage,
		Declared:        true,
		Initialized:     true,
		DeclarationLine: node.Line,
	}, node)
}

func (a *Analyzer) visitFuncDecl(node *parser.ASTNode) {
	name := fieldValue(node, "name")
	a.enterScope(name)
	defer a.leaveScope()

	if params := childByName(node, "Parameters"); params != nil {
		for _, param := range params.Children {
			a.declare(&Symbol{
				Name:            fieldValue(param, "name"),
				Type:            fieldValue(param, "type"),
				Kind:            SymbolParameter,
				Declared:        true,
				Initialized:     true,
				DeclarationLine: param.Line,
			}, param)
		}
	}

	if block := childByName(node, "Block"); block != nil {
		a.visitBlock(block, false)
	}
}

func (a *Analyzer) visitBlock(node *parser.ASTNode, nested bool) {
	if nested {
		a.enterScope(a.scope.name + ".block")
		defer a.leaveScope()
	}

	for _, child := range node.Children {
		a.visitStatement(child)
	}
}

func (a *Analyzer) visitStatement(node *parser.ASTNode) {
	switch node.Name {
	case "VarDecl":
		a.visitVarDecl(node)
	case "AssignStmt":
		a.visitAssignStmt(node)
	case "ExprStmt", "ReturnStmt":
		for _, child := range node.Children {
			a.evalExpr(child)
		}
	case "IfStmt":
		a.visitIfStmt(node)
	case "Block":
		a.visitBlock(node, true)
	}
}

func (a *Analyzer) visitVarDecl(node *parser.ASTNode) {
	name := fieldValue(node, "name")
	declaredType := fieldValue(node, "type")
	symbol := &Symbol{
		Name:            name,
		Type:            declaredType,
		Kind:            SymbolVariable,
		Declared:        true,
		DeclarationLine: node.Line,
	}

	value := childByName(node, "value")
	if value != nil {
		actual := a.evalExpr(value)
		symbol.Initialized = true
		if actual.typ != "" && declaredType != "" && !isAssignableType(declaredType, actual.typ) {
			a.addError("TYPE MISMATCH", fmt.Sprintf("нельзя присвоить значение типа %s переменной %s типа %s", actual.typ, name, declaredType), node)
		}
		a.addTriad("=", name, actual.place)
	}

	a.declare(symbol, node)
}

func (a *Analyzer) visitAssignStmt(node *parser.ASTNode) {
	name := fieldValue(node, "left")
	operator := fieldValue(node, "operator")
	right := a.evalExpr(childByName(node, "right"))

	if operator == ":=" {
		if _, exists := a.scope.symbols[name]; exists {
			a.addError("REDECLARATION", fmt.Sprintf("переменная %s уже объявлена в области %s", name, a.scope.name), node)
			return
		}
		a.declare(&Symbol{
			Name:            name,
			Type:            right.typ,
			Kind:            SymbolVariable,
			Declared:        true,
			Initialized:     true,
			DeclarationLine: node.Line,
		}, node)
		a.addTriad("=", name, right.place)
		return
	}

	symbol := a.lookup(name)
	if symbol == nil {
		a.addError("UNDECLARED VARIABLE", fmt.Sprintf("переменная %s используется до объявления", name), node)
		return
	}
	if symbol.Kind != SymbolVariable && symbol.Kind != SymbolParameter {
		a.addError("INVALID ASSIGNMENT", fmt.Sprintf("идентификатор %s не является переменной", name), node)
		return
	}
	if right.typ != "" && symbol.Type != "" && !isAssignableType(symbol.Type, right.typ) {
		a.addError("TYPE MISMATCH", fmt.Sprintf("нельзя присвоить значение типа %s переменной %s типа %s", right.typ, name, symbol.Type), node)
	}
	symbol.Initialized = true
	a.addTriad("=", name, right.place)
}

func (a *Analyzer) visitIfStmt(node *parser.ASTNode) {
	condition := a.evalExpr(childByName(node, "condition"))
	if condition.typ != "" && condition.typ != "bool" {
		a.addError("TYPE MISMATCH", "условие оператора if должно иметь тип bool", node)
	}
	if thenBlock := childByName(node, "then"); thenBlock != nil && len(thenBlock.Children) > 0 {
		a.visitBlock(thenBlock.Children[0], true)
	}
	if elseBlock := childByName(node, "else"); elseBlock != nil && len(elseBlock.Children) > 0 {
		a.visitBlock(elseBlock.Children[0], true)
	}
}

func (a *Analyzer) evalExpr(node *parser.ASTNode) exprResult {
	if node == nil {
		return exprResult{}
	}

	switch node.Name {
	case "value", "right", "left", "condition", "callee", "object", "GroupedExpr":
		if len(node.Children) == 0 {
			return exprResult{}
		}
		return a.evalExpr(node.Children[0])
	case "IntLiteral":
		return exprResult{typ: "int", place: fieldValue(node, "value")}
	case "RealLiteral":
		return exprResult{typ: "float64", place: fieldValue(node, "value")}
	case "BoolLiteral":
		return exprResult{typ: "bool", place: fieldValue(node, "value")}
	case "StringLiteral":
		return exprResult{typ: "string", place: fieldValue(node, "value")}
	case "Identifier":
		name := fieldValue(node, "name")
		symbol := a.lookup(name)
		if symbol == nil {
			a.addError("UNDECLARED VARIABLE", fmt.Sprintf("идентификатор %s используется до объявления", name), node)
			return exprResult{place: name}
		}
		return exprResult{typ: symbol.Type, place: name}
	case "UnaryExpr":
		return a.evalUnary(node)
	case "BinaryExpr":
		return a.evalBinary(node)
	case "SelectorExpr":
		return a.evalSelector(node)
	case "CallExpr":
		return a.evalCall(node)
	}

	return exprResult{}
}

func (a *Analyzer) evalUnary(node *parser.ASTNode) exprResult {
	operator := fieldValue(node, "operator")
	operand := exprResult{}
	if len(node.Children) > 0 {
		operand = a.evalExpr(node.Children[0])
	}
	switch operator {
	case "!":
		if operand.typ != "" && operand.typ != "bool" {
			a.addError("TYPE MISMATCH", "оператор ! применим только к bool", node)
		}
		ref := a.addTriad(operator, operand.place, "")
		return exprResult{typ: "bool", place: ref}
	case "-":
		if operand.typ != "" && !isNumericType(operand.typ) {
			a.addError("TYPE MISMATCH", "унарный минус применим только к числовым типам", node)
		}
		ref := a.addTriad(operator, operand.place, "")
		return exprResult{typ: operand.typ, place: ref}
	}
	return operand
}

func (a *Analyzer) evalBinary(node *parser.ASTNode) exprResult {
	operator := fieldValue(node, "operator")
	left := a.evalExpr(childByName(node, "left"))
	right := a.evalExpr(childByName(node, "right"))

	resultType := ""
	switch operator {
	case "+", "-", "*", "/", "%":
		if left.typ != "" && right.typ != "" && (!isNumericType(left.typ) || !isNumericType(right.typ)) {
			a.addError("TYPE MISMATCH", fmt.Sprintf("оператор %s применим только к числовым типам", operator), node)
		}
		resultType = numericResultType(left.typ, right.typ)
	case "&&", "||":
		if left.typ != "" && right.typ != "" && (left.typ != "bool" || right.typ != "bool") {
			a.addError("TYPE MISMATCH", fmt.Sprintf("оператор %s применим только к bool", operator), node)
		}
		resultType = "bool"
	case "==", "!=", "<", "<=", ">", ">=":
		if left.typ != "" && right.typ != "" && !isComparableTypes(left.typ, right.typ) {
			a.addError("TYPE MISMATCH", fmt.Sprintf("нельзя сравнить значения типов %s и %s", left.typ, right.typ), node)
		}
		resultType = "bool"
	}

	ref := a.addTriad(operator, left.place, right.place)
	return exprResult{typ: resultType, place: ref}
}

func (a *Analyzer) evalSelector(node *parser.ASTNode) exprResult {
	object := a.evalExpr(childByName(node, "object"))
	field := fieldValue(node, "field")
	if object.typ == "package" {
		return exprResult{typ: "function", place: object.place + "." + field}
	}
	return exprResult{typ: "selector", place: object.place + "." + field}
}

func (a *Analyzer) evalCall(node *parser.ASTNode) exprResult {
	callee := a.evalExpr(childByName(node, "callee"))
	args := childByName(node, "Arguments")
	places := make([]string, 0)
	if args != nil {
		for _, arg := range args.Children {
			places = append(places, a.evalExpr(arg).place)
		}
	}
	ref := a.addTriad("call", callee.place, strings.Join(places, ", "))
	return exprResult{typ: "void", place: ref}
}

func (a *Analyzer) declare(symbol *Symbol, node *parser.ASTNode) {
	if symbol.Name == "" {
		return
	}
	if _, exists := a.scope.symbols[symbol.Name]; exists {
		a.addError("REDECLARATION", fmt.Sprintf("идентификатор %s уже объявлен в области %s", symbol.Name, a.scope.name), node)
		return
	}
	symbol.Scope = a.scope.name
	if symbol.DeclarationLine == 0 {
		symbol.DeclarationLine = node.Line
	}
	a.scope.symbols[symbol.Name] = symbol
	a.symbols = append(a.symbols, symbol)
}

func (a *Analyzer) lookup(name string) *Symbol {
	for current := a.scope; current != nil; current = current.parent {
		if symbol, ok := current.symbols[name]; ok {
			return symbol
		}
	}
	return nil
}

func (a *Analyzer) enterScope(name string) {
	if name == "" {
		name = "block"
	}
	a.scope = &scope{name: name, parent: a.scope, symbols: make(map[string]*Symbol)}
}

func (a *Analyzer) leaveScope() {
	if a.scope.parent != nil {
		a.scope = a.scope.parent
	}
}

func (a *Analyzer) addTriad(operation, arg1, arg2 string) string {
	triad := Triad{
		Number:    len(a.triads) + 1,
		Operation: operation,
		Arg1:      arg1,
		Arg2:      arg2,
	}
	a.triads = append(a.triads, triad)
	return fmt.Sprintf("^%d", triad.Number)
}

func (a *Analyzer) addError(errorType, message string, node *parser.ASTNode) {
	line, column := 0, 0
	if node != nil {
		line, column = node.Line, node.Column
	}
	a.errors = append(a.errors, SemanticError{Type: errorType, Message: message, Line: line, Column: column})
}

func functionSymbol(node *parser.ASTNode) *Symbol {
	return &Symbol{
		Name:            fieldValue(node, "name"),
		Type:            "func",
		Kind:            SymbolFunction,
		Declared:        true,
		Initialized:     true,
		DeclarationLine: node.Line,
	}
}

func fieldValue(node *parser.ASTNode, name string) string {
	if node == nil {
		return ""
	}
	for _, field := range node.Fields {
		if field.Name == name {
			return field.Value
		}
	}
	return ""
}

func childByName(node *parser.ASTNode, name string) *parser.ASTNode {
	if node == nil {
		return nil
	}
	for _, child := range node.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

func isAssignableType(expected, actual string) bool {
	return expected == actual || (expected == "float64" && actual == "int")
}

func isComparableTypes(left, right string) bool {
	return isAssignableType(left, right) || isAssignableType(right, left)
}

func isNumericType(typ string) bool {
	return typ == "int" || typ == "float64"
}

func numericResultType(left, right string) string {
	if left == "float64" || right == "float64" {
		return "float64"
	}
	if left == "int" && right == "int" {
		return "int"
	}
	return ""
}
