package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 將明文密碼加密為 bcrypt hash。
// cost 建議使用 bcrypt.DefaultCost (~10) 或更高 (如 12-14)。
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash 比對明文密碼與 bcrypt hash 是否相符。
// 若不符會回傳錯誤。
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
