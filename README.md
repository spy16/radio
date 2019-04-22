# Radio

[![GoDoc](https://godoc.org/github.com/spy16/radio?status.svg)](https://godoc.org/github.com/spy16/radio) [![Go Report Card](https://goreportcard.com/badge/github.com/spy16/radio)](https://goreportcard.com/report/github.com/spy16/radio)

Radio is a `Go` (or `Golang`) library for creating [RESP](https://redis.io/topics/protocol) (**RE**dis **S**erialization **P**rotocol)
compatible services/servers.

## Features

- [Fast](#benchmarks) Redis compatible server library
- Single RESP parser (`radio.Reader`) that can be used for both client-side and server-side parsing
- Parser supports Inline Commands to use with raw tcp clients (example: `telnet`)
- RESP value types to simplify wrapping values and serializing
- RESP Parser that can be used with any `io.Reader` implementation (e.g., AOF files etc.)


## Benchmarks

Benchmarks were run using [redis-benchmark](https://redis.io/topics/benchmarks) tool.

- Go Version: `go version go1.12.1 darwin/amd64`
- Host: `MacBook Pro 15" Intel Core i7, 2.8 GHz, 4 Cores + 16GB Memory`

**Redis**:

- Run Server: `redis-server --port 9736 --appendonly no`
- Run Benchmark: `redis-benchmark -h 127.0.0.1 -p 9736 -q -t PING -c 100 -n 1000000`

    ```
    PING_INLINE: 80515.30 requests per second
    PING_BULK: 78678.20 requests per second
    ```

**Redcon**:

- Run Server: See [tidwall/redcon](https://github.com/tidwall/redcon#example)
- Except `ping` command, everything else was removed from the example above
- Run Benchmark: `redis-benchmark -h 127.0.0.1 -p 6380 -q -t PING -c 100 -n 1000000`

    ```
    PING_INLINE: 71669.17 requests per second
    PING_BULK: 71828.76 requests per second
    ```

**Radio**:

- Run Server: `go run examples/main.go -addr :8080`
- Run Benchmark: `redis-benchmark -h 127.0.0.1 -p 8080 -q -t PING -c 100 -n 1000000`

    ```
    PING_INLINE: 71199.71 requests per second
    PING_BULK: 71301.25 requests per second
    ```


## TODO

- [ ] Add pieplining support
- [ ] Pub sub support
- [ ] Client functions
