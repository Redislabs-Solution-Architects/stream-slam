package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	utils "github.com/Redislabs-Solution-Architects/stream-slam/util"
	"github.com/go-redis/redis/v8"
	"github.com/pborman/getopt/v2"
)

func workerRead(id int, ctx context.Context, redisClient *redis.Client, streamPrefix string, whack bool, count int64, blockms int, focus bool) {
	log.Printf("Starting worker: %d", id)
	sid := id
	if focus {
		sid = 1
	}

	// Try to create a read group and it will fail if already present
	redisClient.XGroupCreateMkStream(
		ctx,
		fmt.Sprintf("%s-%d", streamPrefix, sid),
		fmt.Sprintf("Group-%s", streamPrefix), "0").Err()

	for {
		res, _ := redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    fmt.Sprintf("Group-%s", streamPrefix),
			Consumer: fmt.Sprintf("Consumer-%s-%d", streamPrefix, id),
			Streams:  []string{fmt.Sprintf("%s-%d", streamPrefix, sid), ">"},
			Count:    count,
			Block:    time.Duration(blockms) * time.Second,
		}).Result()
		for _, x := range res {
			for _, y := range x.Messages {
				if whack {
					_, errdel := redisClient.XDel(
						ctx, fmt.Sprintf("%s-%d", streamPrefix, sid),
						y.ID).Result()
					if errdel != nil {
						log.Printf(
							"%s: Unable to DEL message: %s %s ",
							fmt.Sprintf("%s-%d", streamPrefix, sid),
							y.ID,
							errdel)
					}
				} else {
					_, errack := redisClient.XAck(
						ctx, fmt.Sprintf("%s-%d", streamPrefix, sid),
						fmt.Sprintf("Group-%s", streamPrefix),
						y.ID).Result()
					if errack != nil {
						log.Printf(
							"%s: Unable to ack message: %s %s ",
							fmt.Sprintf("%s-%d", streamPrefix, sid),
							y.ID,
							errack)
					}
				}
			}
		}
	}
}

func main() {

	var ctx = context.Background()

	helpFlag := getopt.BoolLong("help", 'h', "display help")

	redisHost := getopt.StringLong("host", 's', "127.0.0.1", "Redis Host")
	redisPassword := getopt.StringLong("password", 'a', "", "Redis Password")
	streamPrefix := getopt.StringLong("stream-prefix", 'x', "stream-slam", "the prefix of the streams created")

	redisPort := getopt.IntLong("port", 'p', 6379, "Redis Port")
	threadCount := getopt.IntLong("threads", 't', 1, "run this many threads")
	whack := getopt.BoolLong("delete-from-stream", 'r', "delete messages from stream")

	count := getopt.Int64Long("count", 'c', 1, "Number of messages to read at a time")
	blockms := getopt.IntLong("block", 'b', 100, "ms to sleep on read")

	focus := getopt.BoolLong("focus", 'f', "Only write to a single stream")

	getopt.Parse()

	if *helpFlag {
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	ctx = context.WithValue(ctx, "host", *redisHost)
	ctx = context.WithValue(ctx, "port", *redisPort)

	var wg sync.WaitGroup

	client := redis.NewClient(&redis.Options{
		Dialer:          utils.RandomDialer, // Randomly pick an IP address from the list of ips retruned
		Password:        *redisPassword,
		Addr:            fmt.Sprintf("%s:%d", *redisHost, *redisPort),
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
		go workerRead(w, ctx, client, *streamPrefix, *whack, *count, *blockms, *focus)
	}
	wg.Wait()

}
