package memcached

import (
	"context"
	"github.com/arazmj/gerdu/cache"
	mc "github.com/arazmj/gomemcached"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	server *mc.Server
}

//Serve start memcached server
func Serve(host string, gerdu cache.UnImplementedCache) *Server {
	server := mc.NewServer(host)
	server.RegisterFunc("get", func(ctx context.Context, req *mc.Request, res *mc.Response) error {
		return getHandler(ctx, req, res, gerdu)
	})
	server.RegisterFunc("gets", func(ctx context.Context, req *mc.Request, res *mc.Response) error {
		return getHandler(ctx, req, res, gerdu)
	})
	server.RegisterFunc("set", func(ctx context.Context, req *mc.Request, res *mc.Response) error {
		return setHandler(ctx, req, res, gerdu)
	})
	server.RegisterFunc("delete", func(ctx context.Context, req *mc.Request, res *mc.Response) error {
		return deleteHandler(ctx, req, res, gerdu)

	})
	server.RegisterFunc("incr", func(ctx context.Context, req *mc.Request, res *mc.Response) error {
		return incrHandler(ctx, req, res, gerdu)
	})
	server.RegisterFunc("flush_all", func(ctx context.Context, req *mc.Request, res *mc.Response) error {
		return flushAllHandler(ctx, req, res, gerdu)
	})
	server.RegisterFunc("version", func(ctx context.Context, req *mc.Request, res *mc.Response) error {
		return versionHandler(ctx, req, res, gerdu)
	})
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
	return &Server{server: server}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Stop()
}

var (
	memcachedKeys        sync.Map
	memcachedExpirations sync.Map
)

func getHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {
	for _, key := range req.Keys {
		if expired(key) {
			gerdu.Delete(key)
			memcachedKeys.Delete(key)
			memcachedExpirations.Delete(key)
			continue
		}
		value, ok := gerdu.Get(key)
		if ok {
			log.Printf("Memcached RETRIEVED Key: %s Value: %s\n", key, value)
			res.Values = append(res.Values, mc.Value{Key: key, Flags: "0", Data: []byte(value)})
		}
	}

	res.Response = mc.RespEnd
	return nil
}

func setHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {
	key := req.Key
	value := req.Data
	ttl := ttlFromExptime(req.Exptime)
	if req.Exptime > 0 && ttl <= 0 {
		gerdu.Delete(key)
		memcachedKeys.Delete(key)
		memcachedExpirations.Delete(key)
		res.Response = mc.RespStored
		return nil
	}
	created := gerdu.PutWithTTL(key, string(value), ttl)
	memcachedKeys.Store(key, struct{}{})
	if req.Exptime > 0 {
		memcachedExpirations.Store(key, expirationTime(req.Exptime))
	} else {
		memcachedExpirations.Delete(key)
	}
	if !created {
		log.Printf("Memcached UPDATE Key: %s Value: %s\n", key, value)
	} else {
		log.Printf("Memcached INSERT Key: %s Value: %s\n", key, value)
	}
	res.Response = mc.RespStored
	return nil
}

func ttlFromExptime(exptime int64) time.Duration {
	if exptime <= 0 {
		return 0
	}
	now := time.Now().Unix()
	if exptime <= now {
		return time.Duration(exptime) * time.Second
	}
	return time.Until(time.Unix(exptime, 0))
}

func deleteHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {
	count := 0
	for _, key := range req.Keys {
		if _, exists := gerdu.Get(key); exists {
			ok := gerdu.Delete(key)
			memcachedKeys.Delete(key)
			memcachedExpirations.Delete(key)
			if ok {
				log.Printf("Memcached DELETE Key: %s\n", key)
			}
			count++
		} else {
			log.Printf("Memcached DELETE Key not found: %s\n", key)
		}
	}
	if count > 0 {
		res.Response = mc.RespDeleted
	} else {
		res.Response = mc.RespNotFound
	}
	return nil
}

func incrHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {
	key := req.Key
	increment := req.Value
	var base int64
	if value, exists := gerdu.Get(key); exists {
		var err error
		base, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Printf("Memcached INCREMENT Key %v is not valid \n", key)
			return err
		}
	}

	value := strconv.FormatInt(base+increment, 10)
	log.Printf("Memcached INCREMENTED Key %s value %d to value %s\n", key, req.Value, value)

	gerdu.Put(key, value)

	res.Response = value
	return nil
}

func flushAllHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {
	memcachedKeys.Range(func(key, value interface{}) bool {
		keyString := key.(string)
		gerdu.Delete(keyString)
		memcachedKeys.Delete(keyString)
		memcachedExpirations.Delete(keyString)
		return true
	})
	res.Response = mc.RespOK
	return nil
}

func expirationTime(exptime int64) int64 {
	now := time.Now().Unix()
	if exptime <= now {
		return now + exptime
	}
	return exptime
}

func expired(key string) bool {
	expiresAt, ok := memcachedExpirations.Load(key)
	return ok && time.Now().Unix() >= expiresAt.(int64)
}

func versionHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {
	res.Response = "Gerdu VERSION 0.1"
	return nil
}
