package remoteList

import (
	"fmt"
	"sync"
)

type List struct {
	mu   sync.Mutex
	List []int
	Size int
}

type RemoteList struct {
	Lists []List
	Count int
}

type AppendArgs struct {
	List_ID int
	Value   int
}

type GetArgs struct {
	List_ID int
	Index   int
}

func (remoteList *RemoteList) Append(args AppendArgs, reply *bool) error {
	if args.List_ID < 0 || args.List_ID >= remoteList.Count {
		*reply = false
		return nil
	}

	remoteList.Lists[args.List_ID].mu.Lock()
	defer remoteList.Lists[args.List_ID].mu.Unlock()

	remoteList.Lists[args.List_ID].List = append(remoteList.Lists[args.List_ID].List, args.Value)
	remoteList.Lists[args.List_ID].Size++
	*reply = true
	fmt.Printf("Lista[%d]: %v\n", args.List_ID, remoteList.Lists[args.List_ID].List)
	return nil
}

func (remoteList *RemoteList) Get(args GetArgs, reply *int) error {
	if args.List_ID < 0 || args.List_ID >= remoteList.Count {
		return nil
	}

	remoteList.Lists[args.List_ID].mu.Lock()
	defer remoteList.Lists[args.List_ID].mu.Unlock()

	if args.Index < 0 || args.Index >= len(remoteList.Lists[args.List_ID].List) {
		return nil
	}

	*reply = remoteList.Lists[args.List_ID].List[args.Index]
	return nil
}

func (remoteList *RemoteList) Remove(list_id int, reply *int) error {
	if list_id < 0 || list_id >= remoteList.Count {
		return nil
	}

	remoteList.Lists[list_id].mu.Lock()
	defer remoteList.Lists[list_id].mu.Unlock()

	*reply = remoteList.Lists[list_id].List[len(remoteList.Lists[list_id].List)-1]
	remoteList.Lists[list_id].List = remoteList.Lists[list_id].List[:len(remoteList.Lists[list_id].List)-1]
	remoteList.Lists[list_id].Size--
	fmt.Printf("Lista[%d]: %v\n", list_id, remoteList.Lists[list_id].List)
	return nil
}

func (remoteList *RemoteList) Size(list_id int, reply *int) error {
	if list_id < 0 || list_id >= remoteList.Count {
		return nil
	}

	remoteList.Lists[list_id].mu.Lock()
	defer remoteList.Lists[list_id].mu.Unlock()

	*reply = remoteList.Lists[list_id].Size
	return nil
}

func NewRemoteList() *RemoteList {
	return new(RemoteList)
}
