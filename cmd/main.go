package main

import (
	"flag"
	"log"
	"net"
	"runtime"

	"github.com/drycc-addons/redis-sentinel-proxy/proxy"
	"github.com/redis/go-redis/v9"
)

var (
	listen       = flag.String("listen", ":9999", "listen address")
	master       = flag.String("master", "", "name of the master redis node")
	maxProcs     = flag.Int("max-procs", 1, "sets the maximum number of CPUs that can be executing")
	sentinelAddr = flag.String("sentinel-addr", ":26379", "remote sentinel address")
	sentinelUser = flag.String("sentinel-user", "", "username to use when connecting to the sentinel server.")
	sentinelPass = flag.String("sentinel-pass", "", "password to use when connecting to the sentinel server.")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(*maxProcs)
	listenAddr, err := net.ResolveTCPAddr("tcp", *listen)
	if err != nil {
		log.Fatalf("failed to resolve local address: %s", err)
	}
	options := &redis.Options{
		Addr:     *sentinelAddr,
		Username: *sentinelUser,
		Password: *sentinelPass,
	}
	server := proxy.NewRedisProxyServer(listenAddr, options, *master)
	server.Serve()
}
