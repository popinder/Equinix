package internal

import (
	"math/rand"
	"time"
)

// This is just some code needed for NewFabricClient here in internal/

const allowed_charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789#$&@"

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func CorrelationIdWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func CorrelationId(length int) string {
	return CorrelationIdWithCharset(length, allowed_charset)
}
