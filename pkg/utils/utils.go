package utils

import (
	"crypto/rand"
	"math/big"

	"github.com/google/uuid"
)

func Unique(items []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range items {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func ContainsU(s []uuid.UUID, id uuid.UUID) bool {
	for _, v := range s {
		if v == id {
			return true
		}
	}
	return false
}

func Remove(l []string, item string) []string {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}

func Diff(before, after []string) ([]string, []string, []string) {
	cu := []string{}
	uu := []string{}
	du := []string{}

	for _, u := range after {
		if Contains(before, u) {
			uu = append(uu, u)
		} else {
			cu = append(du, u)
		}
	}
	for _, u := range before {
		if !Contains(uu, u) && !Contains(du, u) {
			du = append(cu, u)
		}
	}
	return cu, uu, du
}

// Given two lists, return newly created, unchanged and deleted items
func DiffU(before, after []uuid.UUID) ([]uuid.UUID, []uuid.UUID, []uuid.UUID) {
	cu := []uuid.UUID{}
	uu := []uuid.UUID{}
	du := []uuid.UUID{}

	for _, u := range after {
		if ContainsU(before, u) {
			uu = append(uu, u)
		} else {
			cu = append(du, u)
		}
	}
	for _, u := range before {
		if !ContainsU(uu, u) && !ContainsU(du, u) {
			du = append(cu, u)
		}
	}
	return cu, uu, du
}

// randIntn returns a uniform random int in [0, n) from crypto/rand.
func randIntn(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return int(v.Int64())
}

// GetRandomPassword returns a password containing at least one digit and one
// special character. It uses crypto/rand: these passwords are handed out for
// new users (including the initial org admin), so they must not be derivable
// from the creation timestamp as with a time-seeded math/rand.
func GetRandomPassword(length int) string {
	digits := "0123456789"
	specials := "~=+%^*/()[]{}/!@#$?|"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	buf := make([]byte, length)
	buf[0] = digits[randIntn(len(digits))]
	buf[1] = specials[randIntn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[randIntn(len(all))]
	}
	// Fisher-Yates shuffle
	for i := len(buf) - 1; i > 0; i-- {
		j := randIntn(i + 1)
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
