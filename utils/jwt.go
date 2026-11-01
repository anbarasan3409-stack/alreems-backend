package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type StudentClaims struct {
	StudentID uint   `json:"student_id"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

type AdminClaims struct {
	AdminID uint   `json:"admin_id"`
	Email   string `json:"email"`
	jwt.RegisteredClaims
}

type TeacherClaims struct {
	TeacherID uint   `json:"teacher_id"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

func getJWTSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable not set")
	}
	return []byte(secret), nil
}

func GenerateRefreshToken(studentID uint, email string) (string, error) {
	jwtSecret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 days
	claims := &StudentClaims{
		StudentID: studentID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateAdminRefreshToken(adminID uint, email string) (string, error) {
	jwtSecret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 days
	claims := &AdminClaims{
		AdminID: adminID,
		Email:   email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateTeacherRefreshToken(teacherID uint, email string) (string, error) {
	jwtSecret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 days
	claims := &TeacherClaims{
		TeacherID: teacherID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateRefreshToken(tokenString string) (*StudentClaims, error) {
	jwtSecret, err := getJWTSecret()
	if err != nil {
		return nil, err
	}

	claims := &StudentClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func ValidateAdminRefreshToken(tokenString string) (uint, error) {
	jwtSecret, err := getJWTSecret()
	if err != nil {
		fmt.Printf("JWT Secret error: %v\n", err)
		return 0, err
	}

	tokenPreview := tokenString
	if len(tokenString) > 20 {
		tokenPreview = tokenString[:20] + "..."
	}
	fmt.Printf("Validating admin token: %s\n", tokenPreview)
	fmt.Printf("JWT Secret loaded: %s\n", string(jwtSecret))

	claims := &AdminClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		fmt.Printf("Token parsing error: %v\n", err)
		return 0, err
	}

	if !token.Valid {
		fmt.Printf("Token is not valid\n")
		return 0, errors.New("invalid admin token")
	}

	fmt.Printf("Token validated successfully for admin ID: %d\n", claims.AdminID)
	return claims.AdminID, nil
}

func ValidateTeacherRefreshToken(tokenString string) (uint, error) {
	jwtSecret, err := getJWTSecret()
	if err != nil {
		return 0, err
	}

	claims := &TeacherClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid teacher token")
	}

	return claims.TeacherID, nil
}
