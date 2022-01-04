package hashid

import (
	"fmt"

	"github.com/speps/go-hashids"
)

var hd *hashids.HashID

func init() {
	hd, _ = hashids.NewWithData(&hashids.HashIDData{
		Alphabet:  "abcdefghijklmnopqrstuvwxyz1234567890",
		Salt:      "***REMOVED***",
		MinLength: 7,
	})
}

// IDFromString returns new RafayID for hash id string
func IDFromString(hashID string) (int64, error) {
	ids, err := hd.DecodeInt64WithError(hashID)
	if err != nil {
		return -1, err
	}

	if len(ids) < 1 {
		return -1, fmt.Errorf("No IDs could be constructed from string %s", hashID)
	}
	return ids[0], nil
}

// HashFromInt64 returns new RafayID for hash id string
func HashFromInt64(id int64) (string, error) {
	stringHash, err := hd.EncodeInt64([]int64{id})
	if err != nil {
		return "", err
	}
	return stringHash, nil
}
