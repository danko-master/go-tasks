// Задачи на слайсы

// type SliceHeader struct {
//     Data uintptr
//     Len  int
//     Cap  int
// }

package main

import (
	"fmt"
	"unsafe"
)

func main() {
	// task6()
	task7()
}

// Что выведет программа?
func task6() {
	// len 0, cap 1000
	// Под капотом выделенного массива все 1000 элементов заполнены дефолтным «нулем» для типа int (то есть 0),
	// но напрямую обратиться к ним по индексу [0] нельзя, пока вы не увеличите len (например, через append или функцией copy).
	slice := make([]int, 0, 1000)

	// len 3, cap 1000
	slice = append(slice, 1, 2, 3)
	// [1, 2, 3]
	fmt.Println(slice)

	// 1. Адрес данных (базового массива)
	fmt.Printf("Адрес данных (&s[0]): %p\n", &slice[0])
	fmt.Printf("Адрес данных (unsafe.SliceData): %p\n", unsafe.SliceData(slice))
	// 2. Адрес заголовка самого слайса
	fmt.Printf("Адрес заголовка (&s): %p\n", &slice)

	process(slice)
	// [0, 0, 0]
	fmt.Println(slice)
	// [0, 0, 0, 0, 0, 0]
	fmt.Println(slice[:6])
}

func process(slice []int) {
	// 1. Адрес данных (базового массива)
	fmt.Printf("process - Адрес данных (&s[0]): %p\n", &slice[0])
	fmt.Printf("process - Адрес данных (unsafe.SliceData): %p\n", unsafe.SliceData(slice))
	// 2. Адрес заголовка самого слайса
	fmt.Printf("process - Адрес заголовка (&s): %p\n", &slice)

	// Адрес данных (&s[0]): 0x3a366c68000
	// Адрес данных (unsafe.SliceData): 0x3a366c68000
	// !!! Адрес заголовка слайса (&s): 0x3a366bde048
	// process - Адрес данных (&s[0]): 0x3a366c68000
	// process - Адрес данных (unsafe.SliceData): 0x3a366c68000
	// !!! process - Адрес заголовка слайса (&s): 0x3a366bde078

	for index := range slice {
		// fmt.Println(slice[index])
		// Адрес исходного массива, он не менялся, поэтому данные отобразятся в исходном слайсе
		slice[index] = 0
	}
}

// Что выведет программа?
func task7() {
	// len 0, cap 1000
	slice := make([]int, 0, 1000)
	slice = append(slice, 1, 2, 3)
	// [1 2 3]
	fmt.Println(slice)
	process7(slice)
	// [1 2 3] - т.к. длина 3
	fmt.Println(slice)
	// [1 2 3 4 5 6] - явно расширим "область видимости"
	fmt.Println(slice[:6])

}

func process7(slice []int) {
	// локальный салйс, не влияет на исходный слайс, НО меняется массив внутри него
	slice = append(slice, 4, 5, 6)
}
