package main

import "fmt"

/*
   Тестовая программа для лабораторной работы
   Здесь есть разные конструкции языка Go
*/

// Функция сложения двух чисел
func add(a int, b int) int {
	return a + b
}

func main() {
	// Объявление переменных
	var x int = 10
	var y int = 5
	var result int

	// Арифметическое выражение
	result = x + y*2

	// Логическое выражение и if-else
	if result > 10 && x != 0 {
		fmt.Println("result больше 10")
	} else {
		fmt.Println("result меньше или равен 10")
	}

	// Цикл for
	for i := 0; i < 3; i++ {
		fmt.Println("i =", i)
	}

	// Вызов функции
	sum := add(x, y)
	fmt.Println("sum =", sum)
}
