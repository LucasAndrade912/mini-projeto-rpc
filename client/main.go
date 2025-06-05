package main

import (
	"fmt"
	remotelist "mini-projeto-rpc/remotelist/pkg"
	"net/rpc"
)

func main() {
	client, err := rpc.Dial("tcp", ":5000")
	if err != nil {
		fmt.Print("dialing:", err)
	}

	var reply bool
	var reply_i int

	fmt.Println("------------ Testando método Append")

	argsAppend := remotelist.AppendArgs{List_ID: 0, Value: 10}
	err = client.Call("RemoteList.Append", argsAppend, &reply)

	if err != nil {
		fmt.Print("Error:", err)
	} else {
		fmt.Println("Elemento adicionado:", reply)
	}

	argsAppend = remotelist.AppendArgs{List_ID: 1, Value: 20}
	err = client.Call("RemoteList.Append", argsAppend, &reply)

	if err != nil {
		fmt.Print("Error:", err)
	} else {
		fmt.Println("Elemento adicionado:", reply)
	}

	argsAppend = remotelist.AppendArgs{List_ID: 2, Value: 30}
	err = client.Call("RemoteList.Append", argsAppend, &reply)

	if err != nil {
		fmt.Print("Error:", err)
	} else {
		fmt.Println("Elemento adicionado:", reply)
	}

	argsAppend = remotelist.AppendArgs{List_ID: 0, Value: 15}
	err = client.Call("RemoteList.Append", argsAppend, &reply)

	if err != nil {
		fmt.Print("Error:", err)
	} else {
		fmt.Println("Elemento adicionado:", reply)
	}

	argsAppend = remotelist.AppendArgs{List_ID: 1, Value: 25}
	err = client.Call("RemoteList.Append", argsAppend, &reply)

	if err != nil {
		fmt.Print("Error:", err)
	} else {
		fmt.Println("Elemento adicionado:", reply)
	}

	argsAppend = remotelist.AppendArgs{List_ID: 0, Value: 50}
	err = client.Call("RemoteList.Append", argsAppend, &reply)

	if err != nil {
		fmt.Print("Error:", err)
	} else {
		fmt.Println("Elemento adicionado:", reply)
	}

	fmt.Println("------------ Testando método Get")

	argsGet := remotelist.GetArgs{List_ID: 0, Index: 1}
	err = client.Call("RemoteList.Get", argsGet, &reply_i)

	if err != nil {
		fmt.Print("Error:", err)
	} else {
		fmt.Println("Elemento obtido:", reply_i)
	}

	argsGet = remotelist.GetArgs{List_ID: 1, Index: 0}
	err = client.Call("RemoteList.Get", argsGet, &reply_i)

	if err != nil {
		fmt.Print("Error:", err)
	} else {
		fmt.Println("Elemento obtido:", reply_i)
	}

	fmt.Println("------------ Testando método Remove")

	err = client.Call("RemoteList.Remove", 0, &reply_i)

	if err != nil {
		fmt.Print("Error:", err)
	} else {
		fmt.Println("Elemento retirado:", reply_i)
	}

	err = client.Call("RemoteList.Remove", 0, &reply_i)

	if err != nil {
		fmt.Print("Error:", err)
	} else {
		fmt.Println("Elemento retirado:", reply_i)
	}
}
