package main

import (
	"flag"
	"fmt"
	"itsplanned/test"
)

func main() {
	handlers := flag.Bool("handlers", false, "Run only handler tests")
	flag.Parse()

	if *handlers {
		fmt.Println("Running handler tests only...")
		test.RunHandlerTests()
	} else {
		fmt.Println("Running all tests...")
		test.RunAllTests()
	}
}
