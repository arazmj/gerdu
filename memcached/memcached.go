package memcached

import (
	"context"
	"github.com/arazmj/gerdu/cache"
	mc "github.com/arazmj/gomemcached"
	log "github.com/sirupsen/logrus"
	"strconv"
)

//Serve start memcached server
func Serve(host string, gerdu cache.UnImplementedCache) {
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
	server.Start()
}

func getHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {
	for _, key := range req.Keys {
		value, ok := gerdu.Get(key)
		if ok {

			log.Printf("Memcached RETRIEVED Key: %s Value: %s\n", key, value)
		}
		res.Values = append(res.Values, mc.Value{Key: key, Flags: "0", Data: []byte(value)})
	}

	res.Response = mc.RespEnd
	return nil
}

func setHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {
	key := req.Key
	value := req.Data
	created := gerdu.Put(key, string(value))
	if !created {
		log.Printf("Memcached UPDATE Key: %s Value: %s\n", key, value)
	} else {
		log.Printf("Memcached INSERT Key: %s Value: %s\n", key, value)
	}
	res.Response = mc.RespStored
	return nil
}

func deleteHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {
	count := 0
	for _, key := range req.Keys {
		if _, exists := gerdu.Get(key); exists {
			ok := gerdu.Delete(key)
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
	log.Printf("Memcached INCREMENTED Key %s value %s to value %d\n", key, req.Value, value)

	gerdu.Put(key, value)

	res.Response = value
	return nil
}

func flushAllHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {

	log.Fatalln("Memcached not implemented flush all")
	res.Response = mc.RespOK
	return nil
}

func versionHandler(ctx context.Context, req *mc.Request, res *mc.Response, gerdu cache.UnImplementedCache) error {
	res.Response = "Gerdu VERSION 0.1"
	return nil
}
