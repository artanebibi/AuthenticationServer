package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

type IHashService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, stored string) bool
}

type HashService struct{}

const saltCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const saltLen = 16
const iterations = 10000

func NewHashService() IHashService {
	return &HashService{}
}

func (h *HashService) generateSalt() (string, error) {
	b := make([]byte, saltLen)
	max := big.NewInt(int64(len(saltCharset)))
	for i := 0; i < saltLen; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = saltCharset[n.Int64()]
	}
	return string(b), nil
}

func (h *HashService) hash(password, salt string, iter int) string {
	combined := []byte(salt + password)
	hash := sha256.Sum256(combined)

	for i := 1; i < iter; i++ {
		hash = sha256.Sum256(hash[:])
	}

	return hex.EncodeToString(hash[:])
}

func (h *HashService) HashPassword(password string) (string, error) {
	salt, err := h.generateSalt()
	if err != nil {
		return "", err
	}

	hash := h.hash(password, salt, iterations)

	return fmt.Sprintf("%s$%d$%s", salt, iterations, hash), nil
}

func (h *HashService) VerifyPassword(password, stored string) bool {
	parts := strings.Split(stored, "$")

	var salt, expectedHash string
	var iter int

	if len(parts) == 2 {
		salt = parts[0]
		expectedHash = parts[1]
		iter = 1
	} else if len(parts) == 3 {
		salt = parts[0]
		_, err := fmt.Sscanf(parts[1], "%d", &iter)
		if err != nil {
			return false
		}
		expectedHash = parts[2]
	} else {
		log.Printf("Invalid stored password format. Parts: %d", len(parts))
		return false
	}

	computed := h.hash(password, salt, iter)

	if len(computed) != len(expectedHash) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(computed), []byte(expectedHash)) == 1
}
