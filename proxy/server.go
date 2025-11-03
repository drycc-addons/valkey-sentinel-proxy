package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/valkey-io/valkey-go"
)

type redisProxyServer struct {
	listener     *net.TCPAddr
	clientOption valkey.ClientOption
	masterName   string
}

func NewRedisProxyServer(listener *net.TCPAddr, clientOption valkey.ClientOption, masterName string) *redisProxyServer {
	return &redisProxyServer{
		listener:     listener,
		clientOption: clientOption,
		masterName:   masterName,
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
	client, err := valkey.NewClient(r.clientOption)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// Use SENTINEL GET-MASTER-ADDR-BY-NAME command to get master node address
	resp := client.Do(context.Background(), client.B().Arbitrary("SENTINEL", "GET-MASTER-ADDR-BY-NAME").Args(r.masterName).Build())
	if err := resp.Error(); err != nil {
		return nil, err
	}

	addr, err := resp.AsStrSlice()
	if err != nil {
		return nil, err
	}

	if len(addr) != 2 {
		return nil, fmt.Errorf("invalid master address response: %v", addr)
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
