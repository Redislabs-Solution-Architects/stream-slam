package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
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

func worker(id int, jobs <-chan int, results chan<- string, redisClient *redis.Client, maxlen int, streamPrefix string) {
	for j := range jobs {
		id, err := redisClient.XAdd(&redis.XAddArgs{
			Stream: fmt.Sprintf("%s-%d", streamPrefix, id),
			MaxLen: int64(maxlen),
			Values: map[string]interface{}{"job": id, "message": j},
		}).Result()
		errHndlr(err)
		results <- id
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
	messageCount := getopt.IntLong("message-count", 'c', 100000, "run this man times")
	maxlen := getopt.IntLong("max-length", 'l', 0, "the capped length of a queue")
	threadCount := getopt.IntLong("threads", 't', 10, "run this many threads")

	getopt.Parse()

	if *helpFlag {
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	rHost = *redisHost
	rPort = *redisPort

	if *maxlen == 0 {
		// throw in an extra 10 for good measure
		*maxlen = *messageCount/(*threadCount) + 10
	}

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

	jobs := make(chan int, *messageCount)
	results := make(chan string, *messageCount)

	for w := 1; w <= *threadCount; w++ {
		go worker(w, jobs, results, client, *maxlen, *streamPrefix)
	}

	for j := 0; j <= *messageCount-1; j++ {
		jobs <- j
	}
	close(jobs)

	// Finally we collect all the results of the work.
	for a := 0; a <= *messageCount-1; a++ {
		<-results
	}
	os.Exit(0)

}
