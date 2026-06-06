package utils

import (
	"crypto/rand"
	"math/big"
)

const defaultAliasLen = 7

var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GenerateAlias(n ...int) string {
	length := defaultAliasLen
	if len(n) > 0 && n[0] > 0 {
		length = n[0]
	}

	b := make([]rune, length)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}
