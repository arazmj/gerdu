package redis

import (
	"crypto/tls"
	"github.com/arazmj/gerdu/cache"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/redcon"
	"strconv"
	"strings"
	"time"
)

func Serve(host string, gerdu cache.UnImplementedCache) {
	go log.Infof("Gerdu started listening Redis at %s", host)
	err := redcon.ListenAndServe(host,
		handleCommands(gerdu),
		handleAccept,
		handleClose,
	)
	if err != nil {
		log.Fatal(err)
	}

}

func handleClose(conn redcon.Conn, err error) {
	log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
}

func handleAccept(conn redcon.Conn) bool {
	log.Printf("accept: %s", conn.RemoteAddr())
	return true
}

func ServeTLS(host string, tlsCert, tlsKey string, gerdu cache.UnImplementedCache) {
	go log.Infof("Gerdu started listening Redis at %s", host)
	certificate, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	err = redcon.ListenAndServeTLS(host,
		handleCommands(gerdu),
		func(conn redcon.Conn) bool {
			log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		},

		&tls.Config{Certificates: []tls.Certificate{certificate}},
	)

	if err != nil {
		log.Fatal(err)
	}
}

func parseTTLSeconds(raw []byte) (time.Duration, bool) {
	seconds, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil || seconds <= 0 {
		return 0, false
	}
	return time.Duration(seconds) * time.Second, true
}

func handleCommands(gerdu cache.UnImplementedCache) func(conn redcon.Conn, cmd redcon.Command) {
	return func(conn redcon.Conn, cmd redcon.Command) {
		switch strings.ToLower(string(cmd.Args[0])) {
		default:
			conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
		case "ping":
			conn.WriteString("PONG")
		case "quit":
			conn.WriteString("OK")
			conn.Close()
		case "set":
			if len(cmd.Args) != 3 && len(cmd.Args) != 5 {
				conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
				return
			}
			ttl := time.Duration(0)
			if len(cmd.Args) == 5 {
				if strings.ToLower(string(cmd.Args[3])) != "ex" {
					conn.WriteError("ERR syntax error")
					return
				}
				var ok bool
				ttl, ok = parseTTLSeconds(cmd.Args[4])
				if !ok {
					conn.WriteError("ERR invalid expire time in 'set' command")
					return
				}
			}
			gerdu.PutWithTTL(string(cmd.Args[1]), string(cmd.Args[2]), ttl)
			conn.WriteString("OK")
		case "setex":
			if len(cmd.Args) != 4 {
				conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
				return
			}
			ttl, ok := parseTTLSeconds(cmd.Args[2])
			if !ok {
				conn.WriteError("ERR invalid expire time in 'setex' command")
				return
			}
			gerdu.PutWithTTL(string(cmd.Args[1]), string(cmd.Args[3]), ttl)
			conn.WriteString("OK")
		case "get":
			if len(cmd.Args) != 2 {
				conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
				return
			}
			val, ok := gerdu.Get(string(cmd.Args[1]))
			if !ok {
				conn.WriteNull()
			} else {
				conn.WriteBulk([]byte(val))
			}
		case "del":
			if len(cmd.Args) != 2 {
				conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
				return
			}
			ok := gerdu.Delete(string(cmd.Args[1]))
			if !ok {
				conn.WriteInt(0)
			} else {
				conn.WriteInt(1)
			}
		}
	}
}
