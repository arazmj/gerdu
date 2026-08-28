![Build](https://github.com/arazmj/gerdu/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/arazmj/gerdu)](https://goreportcard.com/report/github.com/arazmj/gerdu)
[![codecov](https://codecov.io/gh/arazmj/gerdu/branch/master/graph/badge.svg)](https://codecov.io/gh/arazmj/gerdu)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/arazmj/gerdu/badges/quality-score.png?b=master)](https://scrutinizer-ci.com/g/arazmj/gerdu/?branch=master)
[![GoDoc](https://godoc.org/github.com/arazmj/gerdu?status.svg)](https://godoc.org/github.com/arazmj/gerdu)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)

![Gerdu](https://github.com/arazmj/gerdu/blob/assets/gerdu_banner.png?raw=true)

## About

Gerdu is a distributed in-memory key-value database server written in [Go](https://go.dev/). It supports LRU, LFU, and weak-reference cache modes, and can serve HTTP, Redis, memcached, and gRPC protocols at the same time.

Gerdu can run as a single node or as a Raft-backed cluster. Prometheus metrics are exposed on the HTTP server.

## Features

- HTTP, Redis, memcached, and gRPC protocol support
- LRU, LFU, and weak-reference cache implementations
- Distributed and fault-tolerant operation through Raft
- Prometheus telemetry at `/metrics`
- Optional TLS for HTTP, gRPC, and Redis
- Graceful Raft leave on interrupt (`Ctrl+C`)

## Requirements

- Go 1.15 or newer, matching the module and current Dockerfile
- Docker, if you want to run Gerdu in a container

TTL support is not exposed in the current `master` API or command-line flags, so keys live until they are evicted by cache policy/capacity or deleted.

## Build

```bash
go build -v .
```

## Docker

The Dockerfile builds Gerdu from the Go 1.15 image and starts the `gerdu` binary.

```bash
docker build -t gerdu .
docker run --rm -p 8080:8080 -p 6379:6379 -p 11211:11211 -p 8081:8081 \
  gerdu -host 0.0.0.0 -protocols redis,mcd,grpc
```

## Usage

```console
Usage of gerdu:
  -capacity string
        Cache capacity. Use K/KB, M/MB, G/GB, or T/TB. (default "64MB")
  -cert string
        SSL certificate public key
  -grpcport int
        gRPC server port number (default 8081)
  -host string
        Host that servers listen on (default "127.0.0.1")
  -httpport int
        HTTP server port number (default 8080)
  -id string
        Node ID (default "master")
  -join string
        Address of an existing Raft node to join
  -key string
        SSL certificate private key
  -log string
        Log level: panic, fatal, error, warn, info, debug, trace (default "info")
  -mcdport int
        memcached server port number (default 11211)
  -protocols string
        Optional protocols: grpc, redis, mcd. Use comma-separated values. HTTP is always enabled.
  -raft string
        Raft bind address (default "127.0.0.1:12000")
  -redisport int
        Redis server port number (default 6379)
  -storage string
        Path to store Raft logs and snapshots; in-memory if not set
  -type string
        Cache type: lru, lfu, weak (default "lru")
```

The HTTP server currently uses Go `net/http` defaults and has no user-facing timeout flags. Raft apply and TCP transport operations use internal 10 second timeouts. On interrupt, Gerdu attempts to leave the Raft cluster before exiting.

## Quick Start

Start Gerdu with all optional protocols enabled:

```bash
gerdu -host 127.0.0.1 -protocols redis,mcd,grpc
```

### HTTP

HTTP is always enabled. Use `/cache/{key}` for key operations.

```bash
curl -i -X PUT --data 'hello' http://127.0.0.1:8080/cache/greeting
curl http://127.0.0.1:8080/cache/greeting
curl -i -X DELETE http://127.0.0.1:8080/cache/greeting
curl -i http://127.0.0.1:8080/cache/greeting
```

### Redis

Enable Redis with `-protocols redis` or include it in the comma-separated list.

```bash
redis-cli -h 127.0.0.1 -p 6379 SET greeting hello
redis-cli -h 127.0.0.1 -p 6379 GET greeting
redis-cli -h 127.0.0.1 -p 6379 DEL greeting
```

Supported Redis commands are `PING`, `QUIT`, `SET`, `GET`, and `DEL`.

### memcached

Enable memcached with `-protocols mcd`.

```bash
printf 'set greeting 0 0 5\r\nhello\r\nget greeting\r\ndelete greeting\r\nquit\r\n' | nc 127.0.0.1 11211
```

Supported memcached commands include `get`, `gets`, `set`, `delete`, `incr`, `flush_all`, and `version`.

### gRPC

Enable gRPC with `-protocols grpc`. The service is defined in [`proto/gerdu.proto`](proto/gerdu.proto) and exposes `Put`, `Get`, and `Delete` RPCs.

```bash
grpcurl -plaintext -import-path proto -proto gerdu.proto \
  -d '{"key":"greeting","value":"aGVsbG8="}' \
  127.0.0.1:8081 gerdu.Gerdu/Put

grpcurl -plaintext -import-path proto -proto gerdu.proto \
  -d '{"key":"greeting"}' \
  127.0.0.1:8081 gerdu.Gerdu/Get
```

The gRPC `value` field is bytes, so JSON clients such as `grpcurl` use base64 encoding.

## Distributed Mode

Gerdu can run in single-node or distributed mode. To join an existing cluster, set `-raft`, `-id`, and `-join`.

A 3-node cluster can tolerate one node failure; a 5-node cluster can tolerate two. Use 3 or 5 Raft servers for a practical balance of availability and performance.

```bash
gerdu -id master -httpport 8080 -raft 127.0.0.1:12000
gerdu -id node1  -httpport 8083 -raft 127.0.0.1:12003 -join 127.0.0.1:12000
gerdu -id node2  -httpport 8084 -raft 127.0.0.1:12004 -join 127.0.0.1:12000
```

Nodes can also be added or removed through the HTTP cluster endpoints:

```bash
curl -X POST http://127.0.0.1:8080/join \
  -H 'content-type: application/json' \
  -d '{"id":"node3","addr":"127.0.0.1:12005"}'

curl -X POST --data 'node3' http://127.0.0.1:8080/leave
```

## Telemetry

Prometheus metrics are available from the HTTP server:

```bash
curl http://127.0.0.1:8080/metrics
```

Example metrics include:

```text
# HELP gerdu_adds_total The total number of new added nodes
# TYPE gerdu_adds_total counter
gerdu_adds_total 52152
# HELP gerdu_deletes_total The total number of deletes nodes
# TYPE gerdu_deletes_total counter
gerdu_deletes_total 23
# HELP gerdu_hits_total The total number of cache hits
# TYPE gerdu_hits_total counter
gerdu_hits_total 1563
# HELP gerdu_misses_total The total number of missed cache hits
# TYPE gerdu_misses_total counter
gerdu_misses_total 16
```

## CI

The Go workflow runs on pushes and pull requests to `master`. It sets up Go, downloads dependencies, builds the project, runs tests, generates coverage excluding examples, and uploads `coverage.txt` to Codecov.

Releases are handled by the GoReleaser workflow on tags and use repository secrets for signing and GitHub publishing.

## Sample applications

Sample applications are available in:

- C# ([HTTP](examples/HTTP/CSharp/CSharp/Program.cs), [gRPC](examples/gRPC/CSharp/CSharpGRPC/Program.cs))
- C++ ([HTTP](examples/HTTP/CPP/main.cpp), [gRPC](examples/gRPC/CPP/main.cpp))
- Dart ([HTTP](examples/HTTP/Dart/bin/Dart.dart))
- Elixir ([HTTP](examples/HTTP/Elixir/lib/go_cache_elixir.ex))
- Erlang ([HTTP](examples/HTTP/Erlang/src/test_gocache.erl))
- GoLang ([HTTP](examples/HTTP/GoLang/main.go), [gRPC](examples/gRPC/GoLang/main.go), [memcached](examples/memcached/GoLang/main.go))
- Groovy ([HTTP](examples/HTTP/Groovy/main.groovy))
- Haskell ([HTTP](examples/HTTP/Haskell/app/Main.hs))
- iOS / Apple platforms ([GerduKit Swift package](examples/iOS/GerduKit))
- Java ([HTTP](examples/HTTP/Java/src/GoCache.java), [gRPC](examples/gRPC/Java/src/main/java/net/amirrazmjou/Main.java))
- Kotlin ([HTTP](examples/HTTP/Kotlin/src/Main.kt))
- NodeJS ([HTTP](examples/HTTP/NodeJS/app.js))
- Objective-C ([HTTP](examples/HTTP/Objective-C/main.m), [gRPC](examples/gRPC/Objective-C/GerduGrpcObjC/main.m))
- Perl ([HTTP](examples/HTTP/Perl/main.pl))
- PHP ([HTTP](examples/HTTP/PHP/test.php))
- Python ([HTTP](examples/HTTP/Python/test.py), [gRPC](examples/gRPC/Python/main.py), [memcached](examples/memcached/Python/test.py))
- R ([HTTP](examples/HTTP/R/main.R))
- Ruby ([HTTP](examples/HTTP/Ruby/go_cache.rb), [gRPC](examples/gRPC/Ruby/main.rb))
- Rust ([HTTP](examples/HTTP/Rust/main.rs), [gRPC](examples/gRPC/Rust/src/main.rs))
- Scala ([HTTP](examples/HTTP/Scala/src/main/scala/com/amirrazmjou/go/cache/example/Example.scala))
- Swift ([HTTP](examples/HTTP/Swift/GoCacheSwift/main.swift))

### iOS and other Apple platforms

iOS apps can talk to Gerdu over its always-on HTTP API. The
[`GerduKit`](examples/iOS/GerduKit) Swift package provides a reusable,
dependency-free client (`put`, `get`, `delete`) with both `async`/`await` and
completion-handler APIs, and runs on iOS, macOS, tvOS, and watchOS:

```swift
import GerduKit

let client = try GerduClient(host: "127.0.0.1", port: 8080)
try await client.put(key: "greeting", value: "hello")
let value = try await client.get(key: "greeting") // "hello"
```

Add it through Swift Package Manager and see [`examples/iOS/GerduKit`](examples/iOS/GerduKit)
for installation, a SwiftUI example, and App Transport Security (HTTP vs TLS)
guidance. For gRPC on Apple platforms, see the
[Objective-C gRPC example](examples/gRPC/Objective-C).
