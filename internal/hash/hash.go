package hash

import "math/rand"

var letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(length int) string {
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		ind := rand.Int31n(int32(len(letters)))
		b[i] = letters[ind]
	}

	return string(b)
}
