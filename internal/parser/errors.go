package parser

import "fmt"

type SyntaxError struct {
	Type     string
	Message  string
	Line     int
	Column   int
	Expected string
	Actual   string
}

func (e SyntaxError) String() string {
	if e.Expected == "" {
		return fmt.Sprintf("Ошибка на строке %d, столбец %d: %s (%s, найдено: %s)",
			e.Line, e.Column, e.Message, e.Type, e.Actual)
	}

	return fmt.Sprintf("Ошибка на строке %d, столбец %d: %s (%s, ожидалось: %s, найдено: %s)",
		e.Line, e.Column, e.Message, e.Type, e.Expected, e.Actual)
}
