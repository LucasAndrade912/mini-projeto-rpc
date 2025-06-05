package main

import (
	"fmt"
	remotelist "mini-projeto-rpc/remotelist/pkg"
	"net"
	"net/rpc"
)

func main() {
	remoteList := remotelist.NewRemoteList()

	remoteList.Lists = make([]remotelist.List, 3)
	remoteList.Count = 3

	for i := 0; i < 3; i++ {
		remoteList.Lists[i].List = make([]int, 0)
		remoteList.Lists[i].Size = 0
		fmt.Printf("List %d initialized.\n", i)
	}

	rpcs := rpc.NewServer()
	rpcs.Register(remoteList)
	l, e := net.Listen("tcp", "[localhost]:5000")

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
