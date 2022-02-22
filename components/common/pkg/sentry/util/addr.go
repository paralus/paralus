package util

import (
	"strconv"
	"strings"
)

// ParseAddr parses addr into host and port
func ParseAddr(addr string) (host string, port int) {
	idx := strings.Index(addr, ":")
	if idx >= 0 {
		host = addr[0:idx]
		p, _ := strconv.ParseInt(addr[idx+1:], 10, 64)
		port = int(p)
	}
	return
}
