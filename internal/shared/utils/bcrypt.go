package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	cost := bcrypt.DefaultCost

	hash, hashErr := bcrypt.GenerateFromPassword([]byte(password), cost)
	if hashErr != nil {
		log.Println("Hash error: ", hashErr)
		return "", hashErr
	}
	return string(hash), nil
}

func CompareHashAndPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateRandomStringForHashing(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		log.Println("Error generating random string", err)
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
