package logging

import (
	"fmt"
	"log"
	"time"
)

// Infoln - prints a line to the log and to the console
func Infoln(msg string) {
	fmt.Printf("%s ", time.Now().Format(time.RFC3339))
	fmt.Println(msg)
	log.Printf("%s ", time.Now().Format(time.RFC3339))
	log.Println(msg)
}

// Infof - prints a formatted message to the log and to the console
func Infof(msg string, args ...interface{}) {
	fmt.Printf("%s ", time.Now().Format(time.RFC3339))
	fmt.Printf(msg, args...)
	log.Printf("%s ", time.Now().Format(time.RFC3339))
	log.Printf(msg, args...)
}

// Fatal - prints a formatted message to the log and to the console
func Fatal(args ...interface{}) {
	fmt.Printf("%s ", time.Now().Format(time.RFC3339))
	fmt.Printf("Fatal error: %s", args)
	log.Printf("%s ", time.Now().Format(time.RFC3339))
	log.Fatal(args...)
}

// Panic - prints a formatted message to the log and to the console
func Panic(err error) {
	fmt.Printf("%s ", time.Now().Format(time.RFC3339))
	fmt.Printf("Fatal error: %s", fmt.Sprint(err))
	log.Printf("%s ", time.Now().Format(time.RFC3339))
	log.Fatal(fmt.Sprint(err))
}
