// This is a helper for generating bcrypt hashes for passwords manually.
// Example: Admin needs to be created in the DB with a hashed password manually
// CLI: go run hash.go "Admin#123" -> prints the hash to use in SQL insert

package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run hash.go <password>")
		return
	}
	pwd := os.Args[1]
	h, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(h))
}
