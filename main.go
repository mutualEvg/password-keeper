// Package main is a placeholder file.
// Use cmd/server/main.go to run the server or cmd/client/main.go to run the client.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("GophKeeper - Secure Password Manager")
	fmt.Println("")
	fmt.Println("To run the server:")
	fmt.Println("  go run cmd/server/main.go")
	fmt.Println("  or: make server && ./bin/server")
	fmt.Println("")
	fmt.Println("To run the client:")
	fmt.Println("  go run cmd/client/main.go [command]")
	fmt.Println("  or: make client && ./bin/client [command]")
	fmt.Println("")
	fmt.Println("For more information, see README.md")
}
