package util

import (
	"math"
	"strconv"
	"strings"
)

// ParseAddr parses addr into host and port
func ParseAddr(addr string) (host string, port int32) {
	idx := strings.Index(addr, ":")
	if idx >= 0 {
		host = addr[0:idx]
		p, _ := strconv.ParseInt(addr[idx+1:], 10, 64)
		if p > 0 && p <= math.MaxInt32 {
			port = int32(p)
		}
	}
	return
}
