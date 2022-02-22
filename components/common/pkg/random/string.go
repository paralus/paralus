package random

import (
	"math/rand"
	"strings"
	"time"
)

var (
	_src = rand.NewSource(time.Now().UnixNano())
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// NewRandomString returns random string of given size
func NewRandomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, _src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = _src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func NewLowerRandomString(n int) string {
	s := NewRandomString(n)
	return strings.ToLower(s)
}
