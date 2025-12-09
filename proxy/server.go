package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/valkey-io/valkey-go"
)

type redisProxyServer struct {
	listener     *net.TCPAddr
	clientOption valkey.ClientOption
	masterName   string
	client       valkey.Client
}

func NewRedisProxyServer(listener *net.TCPAddr, clientOption valkey.ClientOption, masterName string) (*redisProxyServer, error) {
	client, err := valkey.NewClient(clientOption)
	if err != nil {
		return nil, err
	}
	return &redisProxyServer{
		listener:     listener,
		clientOption: clientOption,
		masterName:   masterName,
		client:       client,
	}, nil
}

func (r *redisProxyServer) proxy(local io.ReadWriteCloser, remoteAddr *net.TCPAddr) {
	remote, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		log.Println(err)
		local.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(remote, local)
		remote.CloseWrite()
	}()

	go func() {
		defer wg.Done()
		io.Copy(local, remote)
		local.Close()
	}()

	go func() {
		wg.Wait()
		remote.Close()
	}()
}

func (r *redisProxyServer) master() (*net.TCPAddr, error) {
	// Use SENTINEL GET-MASTER-ADDR-BY-NAME command to get master node address
	resp := r.client.Do(context.Background(), r.client.B().Arbitrary("SENTINEL", "GET-MASTER-ADDR-BY-NAME").Args(r.masterName).Build())
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
			downstream.Close()
			continue
		}
		go r.proxy(downstream, master)
	}
}

func (r *redisProxyServer) Close() error {
	if r.client != nil {
		r.client.Close()
	}
	return nil
}
