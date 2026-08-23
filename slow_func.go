package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

func main() {
	done := make(chan struct{})

	wg := new(sync.WaitGroup)

	// Работает до и после Go 1.25
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	// Имитация работы в 1 сек
	// 	time.Sleep(time.Second)
	// 	// Закрываем канал
	// 	close(done)
	// }()

	// Go 1.25 и позже допустим следующий синтаксис
	// Вызывает wg.Add(1)
	// Запускает ключевое слово go func()
	// Использует defer wg.Done() перед вызовом вашей функции
	wg.Go(func() {
		// Имитация работы в 1 сек
		time.Sleep(time.Second)
		// Закрываем канал
		close(done)
	})

	// Реализация в стандартной библиотеке:
	// func (wg *WaitGroup) Go(f func()) {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		f()
	// 	}()
	// }

	result, err := run(done)
	fmt.Println(result, err)
}

func run(done chan struct{}) (int, error) {
	// Исходные условия - прямой вызов непредскажзуемой функции
	// slowFunction()

	// Меняем решение на работу с каналами
	resCh := make(chan int, 1)
	errCh := make(chan error, 1)

	go func() {
		defer close(resCh)
		defer close(errCh)

		select {
		case v, ok := <-chanSlowFunction():
			if !ok {
				// как вариант errors.New("failed")
				errCh <- fmt.Errorf("failed")
				return
			}

			resCh <- v
		case <-done:
			// как вариант errors.New("timeout")
			errCh <- fmt.Errorf("timeout")
			return
		}
	}()

	return <-resCh, <-errCh
}

// Обертка, для управления состоянием и TTL
func chanSlowFunction() chan int {
	ch := make(chan int)

	// По завершени
	go func() {
		defer close(ch)

		ch <- slowFunction()
	}()

	return ch
}

// Непредсказуемая функция, которую невозможно изменить (например, вызов внешнего сервиса)
func slowFunction() int {
	rnd := rand.N[time.Duration](2000)
	time.Sleep(time.Millisecond * rnd)

	return int(rnd)
}
