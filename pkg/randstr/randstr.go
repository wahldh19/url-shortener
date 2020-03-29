package randstr

import (
	"math/rand"
	"time"
)

// charset contains characters used for generating a random string.
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var randgen = rand.New(rand.NewSource(time.Now().UnixNano()))

// New returns a random string of the requested input length, composed
// by the characters in charset. A zero or negative length input returns
// an empty string. An excessively large requested length may result in
// a panic from make.
func New(length int) string {
	if length <= 0 {
		return ""
	}

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[randgen.Intn(len(charset))]
	}

	return string(b)
}
