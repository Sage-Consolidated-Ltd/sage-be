package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math/big"

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

func HashRandomString(str string) (string, error) {
	cost := 10

	hash, hashErr := bcrypt.GenerateFromPassword([]byte(str), cost)
	if hashErr != nil {
		log.Println("Hash error: ", hashErr)
		return "", hashErr
	}
	return string(hash), nil
}

func Encrypt(plainText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func Decrypt(encryptedText string, key []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
    if err != nil {
        return "", err
    }
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func GenerateSecureOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
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
