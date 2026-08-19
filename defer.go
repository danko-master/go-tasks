package main

import "fmt"

// Что напечатает программа?

// 1 - Оценивается выражение в инструкции return и записывается в возвращаемую ячейку.
// 2 - Выполняются функции defer.
// 3 - Происходит фактический выход из функции (return).

func f1() int {
	// создается локальная переменная x
	x := 1
	defer func() {
		//  увеличивает локальную переменную x на 1 (она станет равна 2).
		x += 1
	}()
	// defer изменяет x после того, как выполняется return x
	// значение x (которое на этот момент равно 1) копируется в специальную ячейку для возвращаемого значения.
	// Сразу после этого срабатывает defer, меняя x до 2, но возвращаемое значение уже зафиксировано как 1.
	// 1
	return x
}

func f2() (x int) {
	// объявлена в сигнатуре как возвращаемая переменная (x int).
	// присваиваем значение именованной переменной.
	x = 1
	defer func() {
		// увеличивает эту же переменную x на 1 (она становится равна 2).
		x += 1
	}()
	// x — это и есть возвращаемая переменная, изменение в defer напрямую влияет на итоговый результат
	// defer изменяет x после того, как выполняется return, но до того, как функция фактически вернет значение
	// 2
	return x
}

func main() {
	// println(f1())
	// println(f2())
	// dtask1()

	// deferEvaluationAsUsual()
	// deferEvaluationOnDemand()

	// fmt.Println(deferChangeReturnValue().value)

	// Последовательность вывода defer
	// 	defer func() {
	// 		fmt.Println("defer 2")
	// 	}()
	// 	F()
	// // defer 1
	// // 	defer 2
	// // 		panic: panic
	// goroutine 1 [running]:
	// main.F()
	// exit status 2

	// "Контролируемая паника"
	protect(func() {
		panic("Test panic\n")
	})
	fmt.Println("All done")
}

func protect(g func()) {
	defer func() {
		fmt.Println("done")

		// Перехват паники
		// Если паники не было, возвращает nil
		if x := recover(); x != nil {
			// Если паника произошла, перехватывает переданный в panic аргумент.
			fmt.Printf("run time panic: %v", x)
		}
	}()

	fmt.Println("start")
	g()
}

func F() {
	defer func() {
		fmt.Println("defer 1")
	}()

	panic("panic")
}

// В какой последовательности отобразятся цифры
// 3-2-1
func dtask1() {
	defer fmt.Println(1)
	defer fmt.Println(2)
	defer fmt.Println(3)
}

type data struct {
	value int
}

func deferEvaluationAsUsual() {
	d := data{}
	// 0
	// Аргументы функции для defer вычисляются в момент объявления, т.е. 0
	defer fmt.Println(d.value)

	d.value++
}

func deferEvaluationOnDemand() {
	d := data{}
	// Чтобы значение было то, которое будет в конце основной функции, т.е. 1
	defer func() {
		// 1
		fmt.Println(d.value)
	}()

	d.value++
}

func deferChangeReturnValue() (d data) {
	defer func() {
		// Будет возвращено 43
		d.value = 43
	}()

	d.value = 42
	return
}
