package test

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Run tests
	fmt.Println("Running all tests...")
	code := m.Run()
	fmt.Println("Tests completed.")

	// Exit with the test status code
	os.Exit(code)
}
