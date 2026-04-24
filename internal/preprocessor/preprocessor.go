package preprocessor

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

func Process(input string) (string, []string, error) {
	var messages []string

	// Проверка на незакрытый многострочный комментарий
	openPos := strings.Index(input, "/*")
	closePos := strings.Index(input, "*/")

	if openPos == -1 && closePos != -1 {
		return "", messages, errors.New("обнаружено закрытие многострочного комментария без открытия")
	}
	if openPos != -1 && closePos == -1 {
		return "", messages, errors.New("обнаружен незакрытый многострочный комментарий")
	}
	if openPos != -1 && closePos != -1 && closePos < openPos {
		return "", messages, errors.New("обнаружено закрытие многострочного комментария без открытия")
	}

	// Регулярные выражения
	reMulti := regexp.MustCompile(`(?s)/\*.*?\*/`)
	reSingle := regexp.MustCompile(`//.*`)
	reSpaces := regexp.MustCompile(`[ \t]+`)

	// Подсчёт комментариев
	multiCount := len(reMulti.FindAllString(input, -1))
	singleCount := len(reSingle.FindAllString(input, -1))

	// Удаление комментариев
	result := reMulti.ReplaceAllString(input, "")
	result = reSingle.ReplaceAllString(result, "")

	// Разбиваем по строкам
	lines := strings.Split(result, "\n")
	cleanedLines := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		line = reSpaces.ReplaceAllString(line, " ")
		cleanedLines = append(cleanedLines, line)
	}

	finalResult := strings.Join(cleanedLines, "\n")

	messages = append(messages, "Удалено многострочных комментариев: "+strconv.Itoa(multiCount))
	messages = append(messages, "Удалено однострочных комментариев: "+strconv.Itoa(singleCount))
	messages = append(messages, "Итоговое количество строк: "+strconv.Itoa(len(cleanedLines)))

	return finalResult, messages, nil
}
