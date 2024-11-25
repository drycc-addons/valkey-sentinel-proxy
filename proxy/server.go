package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/redis/go-redis/v9"
)

type redisProxyServer struct {
	listener    *net.TCPAddr
	sentinel    *redis.Options
	master_name string
}

func NewRedisProxyServer(listener *net.TCPAddr, sentinel *redis.Options, master_name string) *redisProxyServer {
	return &redisProxyServer{
		listener:    listener,
		sentinel:    sentinel,
		master_name: master_name,
	}
}

func (r *redisProxyServer) proxy(local io.ReadWriteCloser, remoteAddr *net.TCPAddr) {
	remote, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		log.Println(err)
		local.Close()
		return
	}
	go func(r io.Reader, w io.WriteCloser) {
		io.Copy(w, r)
		w.Close()
	}(local, remote)
	go func(r io.Reader, w io.WriteCloser) {
		io.Copy(w, r)
		w.Close()
	}(remote, local)
}

func (r *redisProxyServer) master() (*net.TCPAddr, error) {
	sentinel := redis.NewSentinelClient(r.sentinel)
	addr, err := sentinel.GetMasterAddrByName(context.Background(), r.master_name).Result()
	if err != nil {
		return nil, err
	}
	redisMasterAddr, err := net.ResolveTCPAddr(
		"tcp",
		fmt.Sprintf("%s:%s", addr[0], addr[1]),
	)
	if err != nil {
		return nil, err
	}
	return redisMasterAddr, nil
}

func (r *redisProxyServer) Serve() {
	listener, err := net.ListenTCP("tcp", r.listener)
	if err != nil {
		log.Fatal(err)
	}

	for {
		downstream, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		master, err := r.master()
		if err != nil {
			log.Println(err)
			continue
		}
		go r.proxy(downstream, master)
	}
}
