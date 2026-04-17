package main

import (
	"fmt"
	"os"

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

	fmt.Println("=== Сообщения ===")
	if len(messages) == 0 {
		fmt.Println("Ошибок не выявлено")
	} else {
		for _, msg := range messages {
			fmt.Println(msg)
		}
	}
}
