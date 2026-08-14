// Задачи на concurrency (параллельное выполнение)

package main

import (
	"runtime"
	"sync"
)

func main() {
	// task1()
	// task1_solution1()
	task2()
}

// Что выведет программа?
func task1() {
	for _, val := range []int{1, 2, 3} {
		// Не успеет отработать, т.к. основная программа завершится раньше
		go println(val)
	}
}

// Как исправить?
func task1_solution1() {
	var wg sync.WaitGroup
	wg.Add(3) // Увеличиваем счетчик горутин
	for _, val := range []int{1, 2, 3} {
		go func(v int) {
			defer wg.Done() // Уменьшаем счетчик при выходе из горутины
			println(val)
		}(val) // Передаем значение аргументом, чтобы избежать гонки данных
	}

	wg.Wait() // Ждем, пока счетчик не станет равным 0
}

// Что выведет программа?
func task2() {
	ch := make(chan string)

	// fatal error: all goroutines are asleep - deadlock!
	// Случай 1:
	// Этот код заставляет текущую горутину (легковесный поток в Go) остановиться и ждать,
	// пока другая горутина не пришлет данные в этот канал.
	// text := <-ch
	// println("Hello, ", text)

	// Закрыли канал
	close(ch)
	go func() {
		// Должны ждать zero value ""
		text := <-ch
		// Выведет "Hello,  "
		println("Hello, ", text)
	}()
	// Сборщик мусора заблокирует горутину, которая его вызвала, до завершения сборки мусора
	runtime.GC() // Принудительный запуск сборки мусора. В данном случае GC выступает в роли задержки для горутины main.

	// Увидим "Hello,  "
}
