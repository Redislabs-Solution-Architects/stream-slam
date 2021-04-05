package utils

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strings"
)

// Random Dialer shared function
func RandomDialer(ctx context.Context, y string, conn string) (net.Conn, error) {
	var ips []net.IP
	x := strings.Split(conn, ":")

	k := net.ParseIP(x[0])
	if k == nil {
		j, reserr := net.LookupIP(x[0])
		for _, z := range j {
			ips = append(ips, z)
		}
		if reserr != nil {
			return nil, reserr
		}
	} else {
		ips = append(ips, k)
	}

	sort.Slice(ips, func(i, j int) bool {
		return bytes.Compare(ips[i], ips[j]) < 0
	})

	n := rand.Int() % len(ips)

	rconn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", ips[n], x[1]))
	return rconn, err
}
