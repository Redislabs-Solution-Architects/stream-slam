## stream-slam

This builds binaries to run stream benchmarking similar to memtier_benchmark.

### Usage

#### slam-writer

- writes messages to either a single or multiple streams
- allows users to set a payload size
- will spread the load over several Redis Enterprise proxies

#### slam-reader

- reads messages to either a single or multiple streams
- optionally allows you to remove messages from the stream


###

```
$ ./slam-reader -h
Usage: slam-reader [-fhr] [-a value] [-b value] [-c value] [-p value] [-s value] [-t value] [-x value] [parameters ...]
 -a, --password=value
                    Redis Password
 -b, --block=value  ms to sleep on read [100]
 -c, --count=value  Number of messages to read at a time [1]
 -f, --focus        Only write to a single stream
 -h, --help         display help
 -p, --port=value   Redis Port [6379]
 -r, --delete-from-stream
                    delete messages from stream
 -s, --host=value   Redis Host [localhost]
 -t, --threads=value
                    run this many threads [1]
 -x, --stream-prefix=value
                    the prefix of the streams created [stream-slam]

```

```
$ ./slam-writer -h
Usage: slam-writer [-fh] [-a value] [-c value] [-l value] [-m value] [-p value] [-s value] [-t value] [-x value] [parameters ...]
 -a, --password=value
                   Redis Password
 -c, --message-count=value
                   run this man times [100000]
 -f, --focus       Only write to a single stream
 -h, --help        display help
 -l, --max-length=value
                   the capped length of a queue
 -m, --msg-size=value
                   number of bytes in the stream message field [1]
 -p, --port=value  Redis Port [6379]
 -s, --host=value  Redis Host [localhost]
 -t, --threads=value
                   run this many threads [1]
 -x, --stream-prefix=value
                   the prefix of the streams created [stream-slam]

```

### Example run to benchmark a single shard

```
./slam-writer -p 10000 -s 127.0.0.1 -c 10000000 -a MYPASS -t 256 -f
```

```
./slam-reader  -p 10000 -s 127.0.0.1  -a MYPASS -t 20 -c 1000 -f
```

Note: when running in focus (single shard mode) you do not need to match the number of threads on the writer and reader.

### Example run to benchmark a multiple shards

```
./slam-writer -p 10000 -s 127.0.0.1 -c 10000000 -a MYPASS -t 256 
```

```
./slam-reader  -p 10000 -s 127.0.0.1  -a MYPASS -t 256 -c 1000 
```


## Building

Require go and make

```
$ make
rm -f slam-writer slam-reader
go build reader/slam-reader.go
go build writer/slam-writer.go
```