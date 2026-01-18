package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)


func hashPassword(password string) (string, error){
   pass, err := bcrypt.GenerateFromPassword([]byte(password), 7)
   return string(pass), err
}

func comparePassword(hashed string, password string) bool{
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

var secretKey = []byte("370f24a916976d4a6642430277d14cb9d9c633b056adeb35a73f672dd2678f6a")

func createToken(email string) (string, error){
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{
		"email": email,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString(secretKey)
	if err != nil{
		return "", err
	}
	return tokenString, nil
}

func verifyToken(tokenString string) error{
	token, err := jwt.Parse(tokenString, func(token *jwt.Token)(interface{}, error){
		return secretKey, nil
	})
	if !token.Valid{
		return fmt.Errorf("Invalid Token")
	}
	return err
}