package cryptoutil

import (
	"encoding/pem"
	"fmt"
)

func decodePEM(pemBytes []byte) (*pem.Block, error) {
	p, _ := pem.Decode(pemBytes)
	if p == nil {
		return nil, fmt.Errorf("invalid pem block")
	}
	return p, nil
}
