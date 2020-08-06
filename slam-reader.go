package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/pborman/getopt/v2"
)

var rHost string
var rPort int

func errHndlr(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func workerRead(id int, redisClient *redis.Client, streamPrefix string) {

	// if the stream does not exist this will fail so we'll wait patiently
	for {
		r, e := redisClient.XRead(&redis.XReadArgs{
			Streams: []string{fmt.Sprintf("%s-%d", streamPrefix, id), "0"},
			Count:   1,
			Block:   100 * time.Millisecond,
		}).Result()

		if len(r) > 0 && e == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	redisClient.XGroupCreate(fmt.Sprintf("%s-%d", streamPrefix, id), fmt.Sprintf("ReadGroup-%d", id), "0").Err()

	for {
		res, _ := redisClient.XReadGroup(&redis.XReadGroupArgs{
			Group:    fmt.Sprintf("ReadGroup-%d", id),
			Consumer: fmt.Sprintf("ReadConsumer-%d", id),
			Streams:  []string{fmt.Sprintf("%s-%d", streamPrefix, id), ">"},
		}).Result()
		for _, x := range res {
			for y := range x.Messages {
				redisClient.Incr(fmt.Sprintf("ReadGroup-%d", id)).Result()
				if y == -1 {
					fmt.Println(y)
				}
			}
		}
	}
}

//func randomDialer(redisHost string, redisPort int) (net.Conn, error) {
func randomDialer() (net.Conn, error) {
	ips, reserr := net.LookupIP(rHost)
	if reserr != nil {
		return nil, reserr
	}

	sort.Slice(ips, func(i, j int) bool {
		return bytes.Compare(ips[i], ips[j]) < 0
	})

	n := rand.Int() % len(ips)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ips[n], rPort))
	return conn, err
}

func main() {

	helpFlag := getopt.BoolLong("help", 'h', "display help")

	redisHost := getopt.StringLong("host", 's', "localhost", "Redis Host")
	redisPassword := getopt.StringLong("password", 'a', "", "Redis Password")
	streamPrefix := getopt.StringLong("stream-prefix", 'x', "stream-slam", "the prefix of the streams created")

	redisPort := getopt.IntLong("port", 'p', 6379, "Redis Port")
	threadCount := getopt.IntLong("threads", 't', 10, "run this many threads")

	getopt.Parse()

	if *helpFlag {
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	rHost = *redisHost
	rPort = *redisPort

	var wg sync.WaitGroup

	client := redis.NewClient(&redis.Options{
		Dialer:          randomDialer, // Randomly pick an IP address from the list of ips retruned
		Password:        *redisPassword,
		DB:              0,
		MinIdleConns:    1,                    // make sure there are at least this many connections
		MinRetryBackoff: 8 * time.Millisecond, //minimum amount of time to try and backupf
		MaxRetryBackoff: 5000 * time.Millisecond,
		MaxConnAge:      0,  //3 * time.Second this will cause everyone to reconnect every 3 seconds - 0 is keep open forever
		MaxRetries:      10, // retry 10 times : automatic reconnect if a proxy is killed
		IdleTimeout:     time.Second,
	})

	wg.Add(*threadCount)

	for w := 1; w <= *threadCount; w++ {
		go workerRead(w, client, *streamPrefix)
	}
	wg.Wait()

}
