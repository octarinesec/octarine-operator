package test_utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString() string {
	return RandomStringWithLength(rand.Intn(10) + 6)
}

func RandomStringWithLength(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandomLabels() map[string]string {
	randomLabels := make(map[string]string)
	for i := 0; i < rand.Intn(10)+2; i++ {
		randomLabels[RandomString()] = RandomString()
	}

	return randomLabels
}
