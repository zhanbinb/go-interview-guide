package main

import "fmt"

func main() {
	var a Animal
	a = Dog{name: "Fido"}
	fmt.Println(a.Speak())

	a = Cat{name: "Garfield"}
	fmt.Println(a.Speak())
}

type Animal interface {
	Speak() string
}

type Dog struct {
	name string
}

func (d Dog) Speak() string {
	return "bark"
}

type Cat struct {
	name string
}

func (c Cat) Speak() string {
	return "meow"
}
