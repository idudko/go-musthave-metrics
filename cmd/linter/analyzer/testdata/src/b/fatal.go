package b

import (
	"log"
	"os"
)

func useLogFatal() {
	log.Fatal("error occurred") // want "log.Fatal should only be used in main.main function"
}

func useOsExit() {
	os.Exit(1) // want "os.Exit should only be used in main.main function"
}

func helperFunction() {
	log.Fatalf("failed: %v", "error") // want "log.Fatal should only be used in main.main function"

	os.Exit(2) // want "os.Exit should only be used in main.main function"
}
