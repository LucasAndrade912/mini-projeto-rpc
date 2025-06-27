package main

import (
	"fmt"
	remotelist "mini-projeto-rpc/remotelist/pkg"
	"net"
	"net/rpc"
)

func main() {
	remoteList := remotelist.NewRemoteList()

	rpcs := rpc.NewServer()
	rpcs.Register(remoteList)
	l, e := net.Listen("tcp", ":5000")

	if e != nil {
		fmt.Println("listen error:", e)
		return
	}

	defer l.Close()
	fmt.Println("Server is listening on [localhost]:5000")
	for {
		conn, err := l.Accept()
		if err == nil {
			go rpcs.ServeConn(conn)
		} else {
			break
		}
	}
}
