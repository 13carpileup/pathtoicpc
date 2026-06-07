package main

import (
	"fmt"
	"math/rand/v2"
	"os"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(n int) string {
	b := make([]byte, n)

	for i := 0; i < n; i++ {
		b[i] = charset[rand.IntN(len(charset))]
	}

	return string(b)
}

func getKey() string {
	return os.Getenv("CF_KEY")
}

func getSig(method string) string {
	prefix := randomString(6)

	unEncoded := fmt.Sprintf("%s/%s", prefix, method)

	return unEncoded

}
