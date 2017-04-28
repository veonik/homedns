package main

import (
	"fmt"
	"os"
)

func debugf(f string, args ...interface{}) {
	if !*verbose {
		return
	}
	fmt.Printf(f, args...)
}

func debugln(s string) {
	if !*verbose {
		return
	}
	fmt.Println(s)
}

func fatalf(f string, args ...interface{}) {
	fmt.Printf(f, args...)
	os.Exit(1)
}

func fatalln(s string) {
	fmt.Println(s)
	os.Exit(1)
}
