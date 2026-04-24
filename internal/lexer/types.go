package lexer

// TokenType представляет тип лексемы
type TokenType string

// Типы лексем
const (
	KEYWORD       TokenType = "KEYWORD"
	IDENTIFIER    TokenType = "IDENTIFIER"
	CONSTANT_INT  TokenType = "CONSTANT_INT"
	CONSTANT_REAL TokenType = "CONSTANT_REAL"
	CONSTANT_STR  TokenType = "CONSTANT_STR"
	CONSTANT_BOOL TokenType = "CONSTANT_BOOL"
	OPERATOR      TokenType = "OPERATOR"
	DELIMITER     TokenType = "DELIMITER"
	ERROR         TokenType = "ERROR"
)

// Token представляет одну лексему
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// LexemeEntry представляет запись в таблице лексем
type LexemeEntry struct {
	ID          int
	Lexeme      string
	Type        TokenType
	Description string
}

// LexerResult содержит результаты лексического анализа
type LexerResult struct {
	Tokens         []Token
	LexemeTable    []LexemeEntry
	ErrorMessages  []string
	SuccessMessage string
}
