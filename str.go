// Задачи на строки

package main

func main() {
	// task3()
	// println(task3_solution1())
	// task4()
	task5()
}

// Что выведет программа?
func task3() string {
	s := "Test"
	// // Ошибка!
	// // cannot assign to s[0] (neither addressable nor a map index expression)
	// // не удается присвоить значение s[0] (не является ни адресуемым, ни индексным выражением карты)
	// // Строка в go неизменяемый тип
	// s[0] = 'R'
	return s
}

func task3_solution1() string {
	s := "Test"
	b := []rune(s)
	// Символьный литерал 'R' — это число: Запись в одинарных кавычках 'R' обозначает числовую константу (код символа 'R', равный 82).
	b[0] = 'R'
	return string(b)
}

// Посчитать кол-во символов в строке
func task4() {
	str := "Привет!"
	charCount := 0

	// Способ 1
	// Преобразуем строку в срез рун и берем его длину
	b := []rune(str)
	charCount = len(b)
	println(charCount)

	if charCount == 7 {
		println("Ok!")
	}

	// Способ 2
	// Инкремент каждой итерации
	charCount = 0
	for range str {
		charCount++
	}
	println(charCount)

	if charCount == 7 {
		println("Second ok!")
	}
}

// Что выведет программа?
func task5() {
	str := "Привет!"
	for i, v := range str {
		// i представляет собой байтовый индекс каждого символа в строке.
		println(i)
		// Дополнительно - как в данном цикле показать каждую букву
		println(string(v))
	}
	// Русские буквы в кодировке UTF-8 кодируются двумя байтами.
	// Получим:
	// 0
	// 2
	// 4
	// 6
	// 8
	// 10
	// 12
}
