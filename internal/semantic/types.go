package semantic

import "fmt"

type SymbolKind string

const (
	SymbolVariable  SymbolKind = "variable"
	SymbolParameter SymbolKind = "parameter"
	SymbolFunction  SymbolKind = "function"
	SymbolPackage   SymbolKind = "package"
)

type Symbol struct {
	Name            string
	Type            string
	Kind            SymbolKind
	Scope           string
	Declared        bool
	Initialized     bool
	DeclarationLine int
}

type Triad struct {
	Number    int
	Operation string
	Arg1      string
	Arg2      string
}

func (t Triad) String() string {
	return fmt.Sprintf("%d: %s(%s, %s)", t.Number, t.Operation, t.Arg1, t.Arg2)
}

type SemanticError struct {
	Type    string
	Message string
	Line    int
	Column  int
}

func (e SemanticError) String() string {
	if e.Line == 0 {
		return fmt.Sprintf("%s: %s", e.Type, e.Message)
	}
	return fmt.Sprintf("Ошибка на строке %d, столбец %d: %s (%s)", e.Line, e.Column, e.Message, e.Type)
}

type Result struct {
	Symbols        []Symbol
	Triads         []Triad
	ErrorMessages  []string
	SuccessMessage string
}
