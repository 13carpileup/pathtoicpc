package backend

import (
	"cmp"
	"crypto/sha512"
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"time"
)

type Param struct {
	key   string
	value string
}

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

func getSig(method string, params []Param) string {
	prefix := randomString(6)

	unEncoded := fmt.Sprintf("%s/%s?", prefix, method)

	params = append(params, Param{key: "apiKey", value: getKey()})
	params = append(params, Param{key: "time", value: fmt.Sprint(time.Now().Unix())})

	slices.SortFunc(params, func(a, b Param) int {
		keycmp := cmp.Compare(a.key, b.key)
		if keycmp != 0 {
			return keycmp
		}
		return cmp.Compare(a.value, b.value)
	})

	for index, value := range params {
		if index != 0 {
			unEncoded = fmt.Sprintf("%s&", unEncoded)
		}

		unEncoded = fmt.Sprintf("%s%s=%s", unEncoded, value.key, value.value)
	}

	unEncoded = fmt.Sprintf("%s#%s", unEncoded, os.Getenv("CF_SECRET"))

	hashBytes := sha512.Sum512([]byte(unEncoded))
	hashString := fmt.Sprintf("%x", hashBytes)

	encoded := fmt.Sprintf("%s%s", prefix, hashString)

	return encoded
}
