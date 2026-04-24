package main

import (
	"fmt"
	"os"
	"strings"

	"compiler_labs/internal/lexer"
	"compiler_labs/internal/preprocessor"
)

func main() {
	inputFile := "examples/test.go"
	outputFile := "output.txt"

	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Println("Ошибка чтения файла:", err)
		return
	}

	// Этап 1: Препроцессинг (ЛР1)
	cleaned, messages, err := preprocessor.Process(string(data))
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	fmt.Println("=== Очищенный код ===")
	fmt.Println(cleaned)

	err = os.WriteFile(outputFile, []byte(cleaned), 0644)
	if err != nil {
		fmt.Println("Ошибка записи результата в файл:", err)
		return
	}

	fmt.Println("\n=== Сообщения препроцессинга ===")
	if len(messages) == 0 {
		fmt.Println("Ошибок не выявлено")
	} else {
		for _, msg := range messages {
			fmt.Println(msg)
		}
	}

	// Этап 2: Лексический анализ (ЛР2)
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("=== ЛЕКСИЧЕСКИЙ АНАЛИЗ (ЛР2) ===")
	fmt.Println(strings.Repeat("=", 50))

	lexer := lexer.NewLexer(cleaned)
	result := lexer.Analyze()

	fmt.Println("\n=== Таблица лексем ===")
	fmt.Println(result.PrintLexemeTable())

	fmt.Println("\n=== Последовательность токенов ===")
	fmt.Println(result.GetTokenSequence())

	fmt.Println("\n=== Результат анализа ===")
	if len(result.ErrorMessages) > 0 {
		fmt.Println("Найдены ошибки:")
		for _, errMsg := range result.ErrorMessages {
			fmt.Println("  - " + errMsg)
		}
	}
	fmt.Println(result.SuccessMessage)
}
