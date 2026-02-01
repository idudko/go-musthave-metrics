package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// Это допустимое использование log.Fatal в main функции
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

	// Это допустимое использование os.Exit в main функции
	if someCondition() {
		os.Exit(1)
	}
}

func someCondition() bool {
	return false
}

// Эта функция может использовать panic в обычном коде (но она будет помечена как плохая)
func recoverPanic() {
	panic("recover") // want "panic should not be used in production code"
}
