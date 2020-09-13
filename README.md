![Build](https://github.com/arazmj/gerdu/workflows/Go/badge.svg)
![Release](https://github.com/arazmj/gerdu/workflows/GoReleaser/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/arazmj/gerdu)](https://goreportcard.com/report/github.com/arazmj/gerdu)
[![codecov](https://codecov.io/gh/arazmj/gerdu/branch/master/graph/badge.svg)](https://codecov.io/gh/arazmj/gerdu)
[![Maintainability](https://api.codeclimate.com/v1/badges/a99a88d28ad37a79dbf6/maintainability)](https://codeclimate.com/github/codeclimate/codeclimate/maintainability)
[![GoDoc](https://godoc.org/github.com/arazmj/gerdu?status.svg)](https://godoc.org/github.com/arazmj/gerdu)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![codebeat badge](https://codebeat.co/badges/05010b5e-17d9-4f5d-a6bb-2c330ff364c8)](https://codebeat.co/projects/github-com-arazmj-gerdu-master)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/arazmj/gerdu/badges/quality-score.png?b=master)](https://scrutinizer-ci.com/g/arazmj/gerdu/?branch=master)

![Gerdu](https://github.com/arazmj/gerdu/blob/assets/gerdu_banner.png?raw=true)

## About
Gerdu is a distributed key-value in-memory database server written in [Go](http://golang.org) programming language.
Currently, it supports two eviction policy [LFU](https://en.wikipedia.org/wiki/Least_frequently_used) and [LRU](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_(LRU)). 
It also supports for weak reference type of cache where the cache consumes as much memory as the garbage collector allows it to use.
<br/>

You can enable [gRPC](https://grpc.io), [redis](https://redis.io) HTTP and [memcached](https://memcached.org) and enjoy taking advantage of all protocols simultaneously.

## Features
- Wire protocol support for Redis and memcached
- Different eviction policy LRU, LFU, weak
- gRPC and HTTP protocol support
- Distributed and fault-tolerant via Raft 
- Telemetry features through Prometheus 

## Build
```bash
go build -v
```

## Usage
```
Usage of gerdu:
  -capacity string
    	The size of cache, once cache reached this capacity old values will evicted.
    	Specify a numerical value followed by one of the following units (not case sensitive)
    	K or KB: Kilobytes
    	M or MB: Megabytes
    	G or GB: Gigabytes
    	T or TB: Terabytes (default "64MB")
  -cert string
    	SSL certificate public key
  -grpcport int
    	the grpc server port number (default 8081)
  -host string
    	The host that server listens (default "127.0.0.1")
  -httpport int
    	the http server port number (default 8080)
  -id string
    	Node ID (default "master")
  -join string
    	Set join address, if any
  -key string
    	SSL certificate private key
  -log string
    	log level can be any of values of 'panic', 'fatal', 'error', 'warn', 'info', 'debug', 'trace' (default "info")
  -mcdport int
    	the memcached server port number (default 11211)
  -protocols string
    	protocol 'grpc', 'redis' or 'mcd' (memcached), multiple comma-separated values, http is not optional
  -raft string
    	Set Raft bind address (default "127.0.0.1:12000")
  -storage string
    	Path to store log files and snapshot, will store in memory if not set
  -type string
    	type of cache, lru or lfu, weak (default "lru")
```

## Example
Example of usage:
To insert or update or delete a key 
```console
$ ./gerdu --protocols=http,grpc,mcd --log trace
INFO[0000] Gerdu started listening HTTP on 127.0.0.1:8080 
INFO[0000] Gerdu started memcached server on 127.0.0.1:11211 
INFO[0000] Gerdu started listening gRPC on 127.0.0.1:8081 

$ curl --request PUT --data '1' http://localhost:8080/cache/1
$ curl --request GET --data '1' http://localhost:8080/cache/1
1
$ curl --request PUT --data '2' http://localhost:8080/cache/2
$ curl --request PUT --data '3' http://localhost:8080/cache/3
$ curl --request PUT --data 'some new value' http://localhost:8080/cache/3
$ curl --request DELETE http://localhost:8080/cache/3 # Delete key 3
$ curl --request GET localhost:8080/cache/3 # Not found 404
```

To retrieve the key
```Bash
$ curl --request GET localhost:8080/cache/3
```

## Distributed Mode
Gerdu can be ran in either single mode or distributed mode. 
You need to specify `--raft`, `--id`, `--join` parameters to join an existing node. <br>
A Gerdu cluster of 3 nodes can tolerate a single node failure, while a cluster of 5 can tolerate 2 node failures. The recommended configuration is to either run 3 or 5 raft servers. This maximizes availability without greatly sacrificing performance.


```Bash
$ ./gerdu -httpport 8083 --join :8080 --id node1 --raft :12003
$ ./gerdu -httpport 8084 --join :8080 --id node2 --raft :12004
```

## Telemetry 
[Prometheus](https://prometheus.io) metrics
```console
$ curl --request GET localhost:8080/metrics
...
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
...
```

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
- Java ([HTTP](examples/HTTP/Java/src/GoCache.java), [gRPC](examples/gRPC/Java/src/main/java/net/amirrazmjou/Main.java))
- Kotlin ([HTTP](examples/HTTP/Kotlin/src/Main.kt))
- NodeJS ([HTTP](examples/HTTP/NodeJS/app.js))
- Objective-C ([HTTP](examples/HTTP/Objective-C/main.m), [gRPC](examples/gRPC/Objective-C/GerduGrpcObjC/main.m))
- Perl ([HTTP](examples/HTTP/Perl/main.pl))
- PHP  ([HTTP](examples/HTTP/PHP/test.php))
- Python ([HTTP](examples/HTTP/Python/test.py), [gRPC](examples/gRPC/Python/main.py), [memcached](examples/memcached/Python/test.py))
- R ([HTTP](examples/HTTP/R/main.R))
- Ruby ([HTTP](examples/HTTP/Ruby/go_cache.rb), [gRPC](examples/gRPC/Ruby/main.rb))
- Rust ([HTTP](examples/HTTP/Rust/main.rs), [gRPC](examples/gRPC/Rust/src/main.rs))
- Scala ([HTTP](examples/HTTP/Scala/src/main/scala/com/amirrazmjou/go/cache/example/Example.scala))
- Swift ([HTTP](examples/HTTP/Swift/GoCacheSwift/main.swift))
