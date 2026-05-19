package semantic

import (
	"fmt"
	"strings"
)

func (r *Result) PrintSymbolTable() string {
	if len(r.Symbols) == 0 {
		return "Таблица символов пуста"
	}

	var sb strings.Builder
	sb.WriteString("Имя\t| Тип\t\t| Вид\t\t| Область\t| Объявлена\t| Инициализирована\t| Строка\n")
	sb.WriteString("--------+---------------+---------------+---------------+---------------+-----------------------+--------\n")
	for _, symbol := range r.Symbols {
		sb.WriteString(fmt.Sprintf(
			"%s\t| %s\t\t| %s\t| %s\t\t| %s\t\t| %s\t\t\t| %d\n",
			symbol.Name,
			symbol.Type,
			symbol.Kind,
			symbol.Scope,
			mark(symbol.Declared),
			mark(symbol.Initialized),
			symbol.DeclarationLine,
		))
	}
	return sb.String()
}

func (r *Result) PrintTriads() string {
	if len(r.Triads) == 0 {
		return "Триады не сформированы"
	}

	var sb strings.Builder
	for _, triad := range r.Triads {
		sb.WriteString(triad.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

func mark(value bool) string {
	if value {
		return "+"
	}
	return "-"
}
