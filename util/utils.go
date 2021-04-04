package utils

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net"
	"sort"
)

// Random Dialer shared function
func RandomDialer(ctx context.Context, f string, x string) (net.Conn, error) {
	ips, reserr := net.LookupIP(fmt.Sprintf("%s", ctx.Value("host")))
	if reserr != nil {
		return nil, reserr
	}

	sort.Slice(ips, func(i, j int) bool {
		return bytes.Compare(ips[i], ips[j]) < 0
	})

	n := rand.Int() % len(ips)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", ips[n], ctx.Value("port")))
	return conn, err
}
