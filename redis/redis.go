package redis

import (
	"crypto/tls"
	"github.com/arazmj/gerdu/cache"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/redcon"
	"strings"
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
			if len(cmd.Args) != 3 {
				conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
				return
			}
			gerdu.Put(string(cmd.Args[1]), string(cmd.Args[2]))
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
