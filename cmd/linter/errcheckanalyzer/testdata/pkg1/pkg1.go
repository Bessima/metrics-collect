package pkg1

import (
	"log"
	"os"
)

// Функция с вызовом panic
func funcWithPanic() {
	panic("error") // want "panic call detected"
}

// Функция с вызовом log.Fatal
func funcWithLogFatal() {
	log.Fatal("error") // want "log.Fatal call outside main.main"
}

// Функция с вызовом log.Fatalf
func funcWithLogFatalf() {
	log.Fatalf("error: %v", "test") // want "log.Fatal call outside main.main"
}

// Функция с вызовом log.Fatalln
func funcWithLogFatalln() {
	log.Fatalln("error") // want "log.Fatal call outside main.main"
}

// Функция с вызовом os.Exit
func funcWithOsExit() {
	os.Exit(1) // want "os.Exit call outside main.main"
}

// Вложенный вызов panic
func nestedPanic() {
	if true {
		panic("nested panic") // want "panic call detected"
	}
}
