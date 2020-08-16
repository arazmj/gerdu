![](https://github.com/arazmj/gerdu/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/arazmj/gerdu)](https://goreportcard.com/report/github.com/arazmj/gerdu)
[![codecov](https://codecov.io/gh/arazmj/gerdu/branch/master/graph/badge.svg)](https://codecov.io/gh/arazmj/gerdu)
[![Maintainability](https://api.codeclimate.com/v1/badges/a99a88d28ad37a79dbf6/maintainability)](https://codeclimate.com/github/codeclimate/codeclimate/maintainability)
[![GoDoc](https://godoc.org/github.com/arazmj/gerdu?status.svg)](https://godoc.org/github.com/arazmj/gerdu)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![codebeat badge](https://codebeat.co/badges/05010b5e-17d9-4f5d-a6bb-2c330ff364c8)](https://codebeat.co/projects/github-com-arazmj-gerdu-master)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/arazmj/gerdu/badges/quality-score.png?b=master)](https://scrutinizer-ci.com/g/arazmj/gerdu/?branch=master)

# Gerdu

## About
Gerdu is a key-value in-memory database server written in Go programming language (http://golang.org/).
Currently, it supports two eviction policy LFU (Least Frequently Used) and LRU (Least Recently Used). 
It also supports for weak reference type of cache where the cache consumes as much memory as the garbage collector allows it to use.
<br/>

You can enable gRPC and HTTP and enjoy taking advantage of both protocols simultaneously. 
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
  -httpport int
        the http server port number (default 8080)
  -key string
        SSL certificate private key
  -protocols string
        protocol grpc or http, multiple values can be selected seperated by comma (default "http")
  -type string
        type of cache, lru or lfu, weak (default "lru")
  -verbose
        verbose logging
```

## Example
Example of usage:
To insert or update a key 
```Bash
curl --request PUT --data '1' http://localhost:8080/cache/1
curl --request PUT --data '2' http://localhost:8080/cache/2
curl --request PUT --data '3' http://localhost:8080/cache/3
curl --request PUT --data 'some new value' http://localhost:8080/cache/3
curl --request DELETE http://localhost:8080/cache/3 # Delete key 3
curl --request GET localhost:8080/cache/3 # Not found 404
```

To retrieve the key
```Bash
curl --request GET localhost:8080/cache/3
```

## Telemetry 
Prometheus metrics
```
curl --request GET localhost:8080/metrics
```

## Sample applications
Sample application s available in:


C++, Dart, Erlang, Groovy, Java, NodeJS, PHP, Python, Ruby, Scala ,C# ,Elixir ,GoLang ,Haskell ,Kotlin ,Objective-C ,Perl ,R ,Rust ,Swift

