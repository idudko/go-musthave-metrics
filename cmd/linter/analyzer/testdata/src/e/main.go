package main

import (
	"log"
)

func helperFunction() {
	log.Fatal("helper error") // want "log.Fatal should only be used in main.main function"
}

func main() {
	// Это допустимое использование log.Fatal в main функции
	if true {
		log.Fatal("main error")
	}
}
