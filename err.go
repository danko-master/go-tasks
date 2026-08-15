package main

// реализовать функцию, которая возвращает ошибку
// запрещено использовать импорт других библиотек

// В языке Go встроенный интерфейс error выглядит предельно просто и содержит всего один метод Error() string
// type error interface {
//     Error() string
// }

// Создаем собственную структуру ошибки
type customError struct {
}

// Реализуем интерфейс error
// Необходимо создать метод Error(), который вернет string
func (ce *customError) Error() string {
	return "Custom error"
}

// Функция handle() имитирует работу обычной функции, которая возвращает ошибку.
// Она возвращает указатель на экземпляр customError.
// Поскольку customError реализует интерфейс error, функция handle() возвращает значение типа error
func handle() error {
	// ...
	return &customError{}
}

// ///////////////////////////////////
// Решение 2
// Объявляем свой тип на основе string
type stringError string

// Реализуем метод Error() string, чтобы тип стал совместим с интерфейсом error
func (e stringError) Error() string {
	return string(e)
}

// Функция, которая возвращает ошибку
func doSomething() error {
	// Конструкция stringError("текст") берет обычную строку и превращает её в значение вашего собственного типа stringError
	// ИмяТипа(значение) используется для приведения базовых типов.
	// «Создай значение типа stringError из этой обычной строки».
	// Только после такого преобразования
	// полученное значение приобретает метод .Error() и становится полноценным error, который может вернуть функция doSomething()
	return stringError("Custom err 2")
}

//////////////////////////////////////

func main() {
	// println(handle())
	err := handle()
	if err != nil {
		println(err.Error())
	}

	println(doSomething().Error())
}
