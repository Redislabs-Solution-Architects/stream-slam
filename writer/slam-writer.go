package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	utils "github.com/Redislabs-Solution-Architects/stream-slam/util"
	"github.com/go-redis/redis/v8"
	"github.com/pborman/getopt/v2"
)

func worker(id int, ctx context.Context, jobs <-chan int, results chan<- string, redisClient *redis.Client, maxlen int, streamPrefix string, msgsize int, focus bool, pipesize int) {
	log.Printf("Starting writer worker: %d\n", id)
	sid := id
	if focus {
		sid = 1
	}
	pipe := redisClient.Pipeline()
	p := 0
	msg := utils.RandStringBytesMaskImprSrcSB(msgsize)
	for j := range jobs {
		id, err := pipe.XAdd(ctx,
			&redis.XAddArgs{
				Stream: fmt.Sprintf("%s-%d", streamPrefix, sid),
				MaxLen: int64(maxlen),
				Values: map[string]interface{}{"job": j, "message": msg, "thread": id},
			}).Result()
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		}
		results <- id
		p += 1
		if p%pipesize == 0 {
			_, pipeerr := pipe.Exec(ctx)
			if pipeerr != nil {
				log.Printf("ERROR flushing pipeline at %d end: %s\n", p, pipeerr)
			}
		}
	}
	_, finalerr := pipe.Exec(ctx)
	if finalerr != nil {
		log.Printf("ERROR flushing pipeline at end: %s\n", finalerr)
	}
}

func main() {

	var ctx = context.Background()

	helpFlag := getopt.BoolLong("help", 'h', "display help")

	redisHost := getopt.StringLong("host", 's', "127.0.0.1", "Redis Host")
	redisPassword := getopt.StringLong("password", 'a', "", "Redis Password")
	streamPrefix := getopt.StringLong("stream-prefix", 'x', "stream-slam", "the prefix of the streams created")

	redisPort := getopt.IntLong("port", 'p', 6379, "Redis Port")
	messageCount := getopt.IntLong("message-count", 'c', 100000, "run this man times")
	maxlen := getopt.IntLong("max-length", 'l', 0, "the capped length of a queue")
	threadCount := getopt.IntLong("threads", 't', 1, "run this many threads")
	msgsize := getopt.IntLong("msg-size", 'm', 1, "number of bytes in the stream message field")
	focus := getopt.BoolLong("focus", 'f', "Only write to a single stream")
	pipesize := getopt.IntLong("pipeline", 'q', 1, "Pipeline size")

	getopt.Parse()

	if *helpFlag {
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	if *maxlen == 0 {
		// throw in an extra 10 for good measure
		*maxlen = *messageCount/(*threadCount) + 10
	}

	client := redis.NewClient(&redis.Options{
		Dialer:          utils.RandomDialer, // Randomly pick an IP address from the list of ips retruned
		Addr:            fmt.Sprintf("%s:%d", *redisHost, *redisPort),
		Password:        *redisPassword,
		DB:              0,
		MinIdleConns:    5,                    // make sure there are at least this many connections
		MinRetryBackoff: 8 * time.Millisecond, //minimum amount of time to try and backupf
		MaxRetryBackoff: 5000 * time.Millisecond,
		MaxConnAge:      0,  //3 * time.Second this will cause everyone to reconnect every 3 seconds - 0 is keep open forever
		MaxRetries:      10, // retry 10 times : automatic reconnect if a proxy is killed
		IdleTimeout:     time.Second,
	})

	jobs := make(chan int, *messageCount)
	results := make(chan string, *messageCount)

	for w := 1; w <= *threadCount; w++ {
		go worker(w, ctx, jobs, results, client, *maxlen, *streamPrefix, *msgsize, *focus, *pipesize)
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
