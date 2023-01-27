package utils

import (
	"math/rand"
	"time"

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

func GetRandomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	specials := "~=+%^*/()[]{}/!@#$?|"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	return string(buf)
}
