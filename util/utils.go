package utils

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
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

func RandStringBytesMaskImprSrcSB(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
