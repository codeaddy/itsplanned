package test

import (
	"fmt"
	"os"
	"os/exec"
)

// RunAllTests runs all tests in the project
func RunAllTests() {
	fmt.Println("Running all tests...")
	cmd := exec.Command("go", "test", "./...", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Tests failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All tests passed!")
}

// RunHandlerTests runs all handler tests
func RunHandlerTests() {
	fmt.Println("Running handler tests...")
	cmd := exec.Command("go", "test", "./handlers/...", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Handler tests failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All handler tests passed!")
}
