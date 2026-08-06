package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

func ComparePassword(hashValue, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashValue), []byte(plain)) == nil
}
