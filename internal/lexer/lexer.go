package lexer

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var keywords = map[string]bool{
	"package": true, "import": true, "func": true, "return": true,
	"var": true, "const": true, "type": true, "struct": true,
	"if": true, "else": true, "switch": true, "case": true,
	"default": true, "for": true, "break": true, "continue": true,
	"range": true, "int": true, "float64": true, "bool": true, "string": true,
}

var operators = map[string]bool{
	"+": true, "-": true, "*": true, "/": true, "%": true,
	"&": true, "|": true, "^": true, "<<": true, ">>": true, "&^": true,
	"+=": true, "-=": true, "*=": true, "/=": true, "%=": true,
	"&=": true, "|=": true, "^=": true, ">>=": true, "<<=": true, "&^=": true,
	"&&": true, "||": true, "!": true, "==": true, "!=": true,
	"<": true, "<=": true, ">": true, ">=": true,
	":=": true, "=": true, "++": true, "--": true, "<-": true,
}

var delimiters = map[string]bool{
	"(": true, ")": true, "[": true, "]": true, "{": true, "}": true,
	";": true, ",": true, ".": true, ":": true,
}

var booleanConstants = map[string]bool{
	"true": true, "false": true,
}

// Lexer выполняет лексический анализ
type Lexer struct {
	input    string
	position int
	line     int
	column   int
	tokens   []Token
	errors   []LexicalError
}

// NewLexer создает новый экземпляр лексера
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:    input,
		position: 0,
		line:     1,
		column:   1,
		tokens:   make([]Token, 0),
		errors:   make([]LexicalError, 0),
	}
}

// Analyze выполняет лексический анализ входного текста
func (l *Lexer) Analyze() *LexerResult {
	for l.position < len(l.input) {
		ch := rune(l.input[l.position])

		// Пропускаем пробелы
		if unicode.IsSpace(ch) {
			if ch == '\n' {
				l.line++
				l.column = 1
			} else {
				l.column++
			}
			l.position++
			continue
		}

		// Проверяем строковые константы
		if ch == '"' || ch == '\'' || ch == '`' {
			l.readString(ch)
			continue
		}

		// Проверяем числа
		if unicode.IsDigit(ch) {
			l.readNumber()
			continue
		}

		// Проверяем идентификаторы и ключевые слова
		if unicode.IsLetter(ch) || ch == '_' {
			l.readIdentifierOrKeyword()
			continue
		}

		// Проверяем операторы и разделители
		if l.isOperatorOrDelimiter(ch) {
			l.readOperatorOrDelimiter()
			continue
		}

		// Неизвестный символ
		l.errors = append(l.errors, NewInvalidCharError(ch, l.line, l.column))
		l.position++
		l.column++
	}

	return l.buildResult()
}

// readString читает строковую константу
func (l *Lexer) readString(quote rune) {
	start := l.position
	startLine := l.line
	startColumn := l.column
	l.position++
	l.column++

	for l.position < len(l.input) {
		ch := rune(l.input[l.position])

		if ch == quote {
			l.position++
			l.column++
			value := l.input[start:l.position]
			l.addToken(CONSTANT_STR, value, startLine, startColumn)
			return
		}

		if ch == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
		l.position++
	}

	// Незакрытая строка
	l.errors = append(l.errors, NewUnclosedStringError(startLine, startColumn))
}

// readNumber читает числовую константу
func (l *Lexer) readNumber() {
	start := l.position
	startLine := l.line
	startColumn := l.column
	hasDecimal := false
	isValid := true

	for l.position < len(l.input) {
		ch := l.input[l.position]

		if unicode.IsDigit(rune(ch)) {
			l.position++
			l.column++
		} else if ch == '.' && !hasDecimal {
			// Проверяем, что это действительно начало дробной части
			if l.position+1 < len(l.input) && unicode.IsDigit(rune(l.input[l.position+1])) {
				hasDecimal = true
				l.position++
				l.column++
			} else {
				break
			}
		} else if ch == '.' && hasDecimal {
			// Две точки подряд - ошибка
			isValid = false
			l.position++
			l.column++
			break
		} else if unicode.IsLetter(rune(ch)) {
			// Буквы в числе - ошибка
			isValid = false
			break
		} else {
			break
		}
	}

	value := l.input[start:l.position]

	if !isValid {
		l.errors = append(l.errors, NewMalformedNumberError(value, startLine, startColumn))
		return
	}

	tokenType := CONSTANT_INT
	if hasDecimal {
		tokenType = CONSTANT_REAL
	}

	l.addToken(tokenType, value, startLine, startColumn)
}

// readIdentifierOrKeyword читает идентификатор или ключевое слово
func (l *Lexer) readIdentifierOrKeyword() {
	start := l.position
	startLine := l.line
	startColumn := l.column

	for l.position < len(l.input) {
		ch := l.input[l.position]
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' {
			l.position++
			l.column++
		} else {
			break
		}
	}

	value := l.input[start:l.position]

	// Проверяем булевы константы
	if booleanConstants[value] {
		l.addToken(CONSTANT_BOOL, value, startLine, startColumn)
		return
	}

	// Проверяем ключевые слова
	if keywords[value] {
		l.addToken(KEYWORD, value, startLine, startColumn)
		return
	}

	// Иначе это идентификатор
	l.addToken(IDENTIFIER, value, startLine, startColumn)
}

// readOperatorOrDelimiter читает оператор или разделитель
func (l *Lexer) readOperatorOrDelimiter() {
	start := l.position
	startLine := l.line
	startColumn := l.column

	// Проверяем на некорректные комментарии (должны быть удалены препроцессором)
	if l.position+1 < len(l.input) {
		twoChar := l.input[start : l.position+2]
		if twoChar == "/*" || twoChar == "*/" {
			l.errors = append(l.errors, LexicalError{
				Type:    "НЕКОРРЕКТНАЯ ЛЕКСЕМА",
				Message: fmt.Sprintf("Комментарий '%s' не должен появляться в очищенном коде. Возможно, ошибка препроцессора.", twoChar),
				Line:    startLine,
				Column:  startColumn,
			})
			l.position += 2
			l.column += 2
			return
		}
	}

	// Пытаемся прочитать двусимвольный оператор
	if l.position+1 < len(l.input) {
		twoChar := l.input[start : l.position+2]
		if operators[twoChar] {
			l.position += 2
			l.column += 2
			l.addToken(OPERATOR, twoChar, startLine, startColumn)
			return
		}
	}

	// Проверяем односимвольные операторы и разделители
	oneChar := string(l.input[l.position])

	if operators[oneChar] {
		l.position++
		l.column++
		l.addToken(OPERATOR, oneChar, startLine, startColumn)
		return
	}

	if delimiters[oneChar] {
		l.position++
		l.column++
		l.addToken(DELIMITER, oneChar, startLine, startColumn)
		return
	}

	// Неизвестный символ
	l.errors = append(l.errors, NewInvalidCharError(rune(l.input[l.position]), startLine, startColumn))
	l.position++
	l.column++
}

// isOperatorOrDelimiter проверяет, является ли символ началом оператора или разделителя
func (l *Lexer) isOperatorOrDelimiter(ch rune) bool {
	s := string(ch)
	if operators[s] || delimiters[s] {
		return true
	}

	// Проверяем двусимвольные операторы
	if l.position+1 < len(l.input) {
		twoChar := l.input[l.position : l.position+2]
		if operators[twoChar] {
			return true
		}
	}

	return false
}

// addToken добавляет токен в список
func (l *Lexer) addToken(tokenType TokenType, value string, line, column int) {
	token := Token{
		Type:   tokenType,
		Value:  value,
		Line:   line,
		Column: column,
	}
	l.tokens = append(l.tokens, token)
}

// buildResult построет окончательный результат анализа
func (l *Lexer) buildResult() *LexerResult {
	result := &LexerResult{
		Tokens:        l.tokens,
		ErrorMessages: make([]string, 0),
	}

	// Добавляем сообщения об ошибках
	for _, err := range l.errors {
		result.ErrorMessages = append(result.ErrorMessages, err.String())
	}

	// Построение таблицы лексем
	result.LexemeTable = l.buildLexemeTable()

	// Успешное сообщение
	if len(result.ErrorMessages) == 0 {
		result.SuccessMessage = "Лексический анализ завершён успешно. " +
			"Обнаружено " + formatNumber(len(l.tokens)) + ". " +
			"Ошибок не найдено."
	} else {
		result.SuccessMessage = "Лексический анализ завершён с ошибками. " +
			"Обнаружено " + formatNumber(len(l.tokens)) + ". " +
			"Найдено ошибок: " + formatNumber(len(result.ErrorMessages)) + "."
	}

	return result
}

// buildLexemeTable построение таблицы лексем
func (l *Lexer) buildLexemeTable() []LexemeEntry {
	entries := make([]LexemeEntry, 0)
	seen := make(map[string]bool)
	id := 1

	for _, token := range l.tokens {
		key := token.Value + "|" + string(token.Type)
		if !seen[key] {
			seen[key] = true
			entry := LexemeEntry{
				ID:     id,
				Lexeme: token.Value,
				Type:   token.Type,
			}

			// Добавляем описание в зависимости от типа
			switch token.Type {
			case KEYWORD:
				entry.Description = getKeywordDescription(token.Value)
			case OPERATOR:
				entry.Description = getOperatorDescription(token.Value)
			case DELIMITER:
				entry.Description = getDelimiterDescription(token.Value)
			default:
				entry.Description = ""
			}

			entries = append(entries, entry)
			id++
		}
	}

	return entries
}

// getKeywordDescription возвращает описание ключевого слова
func getKeywordDescription(keyword string) string {
	descriptions := map[string]string{
		"package":  "Объявление пакета",
		"import":   "Импорт пакетов",
		"func":     "Объявление функции",
		"return":   "Возврат из функции",
		"var":      "Объявление переменной",
		"const":    "Объявление константы",
		"if":       "Условный оператор",
		"else":     "Альтернативный блок условия",
		"for":      "Цикл",
		"switch":   "Оператор выбора",
		"case":     "Вариант в switch",
		"default":  "Вариант по умолчанию",
		"break":    "Выход из цикла",
		"continue": "Переход к следующей итерации",
		"int":      "Целый тип данных",
		"float64":  "Вещественный тип данных",
		"bool":     "Булев тип данных",
		"string":   "Строковый тип данных",
		"range":    "Итерация по диапазону",
		"true":     "Логическое истина",
		"false":    "Логическое ложь",
	}
	if desc, ok := descriptions[keyword]; ok {
		return desc
	}
	return ""
}

// getOperatorDescription возвращает описание оператора
func getOperatorDescription(operator string) string {
	descriptions := map[string]string{
		"+":  "Сложение",
		"-":  "Вычитание",
		"*":  "Умножение",
		"/":  "Деление",
		"%":  "Остаток от деления",
		"=":  "Присваивание",
		":=": "Присваивание с объявлением",
		"==": "Равно",
		"!=": "Не равно",
		"<":  "Меньше",
		">":  "Больше",
		"<=": "Меньше или равно",
		">=": "Больше или равно",
		"&&": "Логическое И",
		"||": "Логическое ИЛИ",
		"!":  "Логическое НЕ",
		"&":  "Побитовое И",
		"|":  "Побитовое ИЛИ",
		"^":  "Побитовое исключающее ИЛИ",
		"<<": "Сдвиг влево",
		">>": "Сдвиг вправо",
		"+=": "Присваивание с добавлением",
		"-=": "Присваивание с вычитанием",
		"*=": "Присваивание с умножением",
		"/=": "Присваивание с делением",
		"++": "Увеличение",
		"--": "Уменьшение",
	}
	if desc, ok := descriptions[operator]; ok {
		return desc
	}
	return ""
}

// getDelimiterDescription возвращает описание разделителя
func getDelimiterDescription(delimiter string) string {
	descriptions := map[string]string{
		"(":  "Открывающая круглая скобка",
		")":  "Закрывающая круглая скобка",
		"[":  "Открывающая квадратная скобка",
		"]":  "Закрывающая квадратная скобка",
		"{":  "Открывающая фигурная скобка",
		"}":  "Закрывающая фигурная скобка",
		";":  "Точка с запятой",
		",":  "Запятая",
		".":  "Точка",
		":":  "Двоеточие",
		"\"": "Двойная кавычка",
		"'":  "Одинарная кавычка",
		"`":  "Обратная кавычка",
	}
	if desc, ok := descriptions[delimiter]; ok {
		return desc
	}
	return ""
}

// formatNumber форматирует число в русский стиль
func formatNumber(n int) string {
	s := "токенов"
	if n%10 == 1 && n%100 != 11 {
		s = "токен"
	} else if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		s = "токена"
	}
	return formatInt(n) + " " + s
}

// formatInt конвертирует число в строку
func formatInt(n int) string {
	return strings.TrimSpace(regexp.MustCompile(`\D`).ReplaceAllString(fmt.Sprintf("%d", n), ""))
}

// PrintTokens выводит токены в формате таблицы
func (r *LexerResult) PrintTokens() string {
	if len(r.Tokens) == 0 {
		return "Токены не найдены"
	}

	var sb strings.Builder
	sb.WriteString("Лексема\t\t| Тип\n")
	sb.WriteString("----------------+----------------------\n")

	for _, token := range r.Tokens {
		sb.WriteString(token.Value + "\t\t| " + string(token.Type) + "\n")
	}

	return sb.String()
}

// GetTokenSequence возвращает последовательность токенов
func (r *LexerResult) GetTokenSequence() string {
	if len(r.Tokens) == 0 {
		return "[]"
	}

	sequence := "["
	for i, token := range r.Tokens {
		if i > 0 {
			sequence += ", "
		}
		sequence += "(" + string(token.Type) + ", " + token.Value + ")"
	}
	sequence += "]"

	return sequence
}

// PrintLexemeTable выводит таблицу лексем
func (r *LexerResult) PrintLexemeTable() string {
	if len(r.LexemeTable) == 0 {
		return "Таблица лексем пуста"
	}

	var sb strings.Builder
	sb.WriteString("ID\t| Лексема\t| Тип\t\t| Описание\n")
	sb.WriteString("----+---------------+---------------+--------------------------------\n")

	for _, entry := range r.LexemeTable {
		sb.WriteString(fmt.Sprintf("%d\t| %s\t| %s\t| %s\n", entry.ID, entry.Lexeme, entry.Type, entry.Description))
	}

	return sb.String()
}
