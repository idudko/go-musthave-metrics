package d

import "log"

func helperFunction() {
	log.Fatal("helper error") // want "log.Fatal should only be used in main.main function"
}
