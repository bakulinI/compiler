package preprocessor

import (
	"errors"
	"regexp"
	"strings"
)

func Process(input string) (string, []string, error) {
	var messages []string

	// Проверка на незакрытый многострочный комментарий
	openCount := strings.Count(input, "/*")
	closeCount := strings.Count(input, "*/")
	if openCount > closeCount {
		return "", messages, errors.New("обнаружен незакрытый многострочный комментарий")
	}

	// Удаление многострочных комментариев: /* ... */
	reMulti := regexp.MustCompile(`(?s)/\*.*?\*/`)
	result := reMulti.ReplaceAllString(input, "")

	// Удаление однострочных комментариев: // ...
	reSingle := regexp.MustCompile(`//.*`)
	result = reSingle.ReplaceAllString(result, "")

	// Разбиваем по строкам
	lines := strings.Split(result, "\n")
	cleanedLines := make([]string, 0)

	for _, line := range lines {
		// Удаляем пробелы и табы в начале/конце строки
		line = strings.TrimSpace(line)

		// Пропускаем пустые строки
		if line == "" {
			continue
		}

		// Сжимаем множественные пробелы до одного
		reSpaces := regexp.MustCompile(`[ \t]+`)
		line = reSpaces.ReplaceAllString(line, " ")

		cleanedLines = append(cleanedLines, line)
	}

	finalResult := strings.Join(cleanedLines, "\n")
	return finalResult, messages, nil
}
