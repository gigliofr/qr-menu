package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Hash che abbiamo messo nel database
	hash := "$2a$10$dopU1ueHFSSkCmD78zuJCe1H0jBgtfnDp.pofNxNrleXL5SEGiCVK"
	
	// Proviamo diverse password
	passwords := []string{"admin", "admin123", "password", "Admin"}
	
	fmt.Println("==============================================")
	fmt.Println("VERIFICA HASH BCRYPT")
	fmt.Println("==============================================")
	fmt.Printf("Hash: %s\n\n", hash)
	
	for _, pwd := range passwords {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
		if err == nil {
			fmt.Printf("✅ Password '%s' CORRETTA!\n", pwd)
		} else {
			fmt.Printf("❌ Password '%s' SBAGLIATA\n", pwd)
		}
	}
	
	fmt.Println("\n==============================================")
	fmt.Println("GENERA NUOVO HASH PER 'admin'")
	fmt.Println("==============================================")
	
	newHash, err := bcrypt.GenerateFromPassword([]byte("admin"), 10)
	if err != nil {
		fmt.Printf("Errore: %v\n", err)
		return
	}
	
	fmt.Printf("Nuovo hash: %s\n", string(newHash))
	
	// Verifica il nuovo hash
	err = bcrypt.CompareHashAndPassword(newHash, []byte("admin"))
	if err == nil {
		fmt.Println("✅ Nuovo hash verificato correttamente!")
	} else {
		fmt.Println("❌ Nuovo hash non funziona")
	}
}
