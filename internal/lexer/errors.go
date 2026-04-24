package lexer

import "fmt"

// LexicalError представляет ошибку лексического анализа
type LexicalError struct {
	Type    string // Тип ошибки
	Message string // Сообщение об ошибке
	Line    int    // Номер строки
	Column  int    // Номер столбца
}

// String возвращает строковое представление ошибки
func (e LexicalError) String() string {
	return fmt.Sprintf("Ошибка на строке %d, столбец %d: %s (%s)", e.Line, e.Column, e.Message, e.Type)
}

// NewInvalidCharError создает ошибку для недопустимого символа
func NewInvalidCharError(char rune, line, column int) LexicalError {
	return LexicalError{
		Type:    "НЕДОПУСТИМЫЙ СИМВОЛ",
		Message: fmt.Sprintf("Недопустимый символ '%c'", char),
		Line:    line,
		Column:  column,
	}
}

// NewMalformedNumberError создает ошибку для некорректного числа
func NewMalformedNumberError(number string, line, column int) LexicalError {
	return LexicalError{
		Type:    "НЕКОРРЕКТНОЕ ЧИСЛО",
		Message: fmt.Sprintf("Некорректное число '%s' (две точки подряд или буквы в числе)", number),
		Line:    line,
		Column:  column,
	}
}

// NewUnclosedStringError создает ошибку для незакрытой строки
func NewUnclosedStringError(line, column int) LexicalError {
	return LexicalError{
		Type:    "НЕЗАКРЫТАЯ СТРОКА",
		Message: "Строковый литерал не закрыт",
		Line:    line,
		Column:  column,
	}
}

// NewInvalidIdentifierError создает ошибку для идентификатора, начинающегося с цифры
func NewInvalidIdentifierError(identifier string, line, column int) LexicalError {
	return LexicalError{
		Type:    "НЕДОПУСТИМЫЙ ИДЕНТИФИКАТОР",
		Message: fmt.Sprintf("Идентификатор не может начинаться с цифры: '%s'", identifier),
		Line:    line,
		Column:  column,
	}
}

// NewUnknownOperatorError создает ошибку для неизвестного оператора
func NewUnknownOperatorError(op string, line, column int) LexicalError {
	return LexicalError{
		Type:    "НЕИЗВЕСТНЫЙ ОПЕРАТОР",
		Message: fmt.Sprintf("Неизвестный оператор '%s'", op),
		Line:    line,
		Column:  column,
	}
}
