package remoteList

import (
	"fmt"
	"os"
	"sync"
)

type List struct {
	mu   sync.Mutex
	List []int
	Size int
}

type RemoteList struct {
	Lists   map[int]*List
	mu      sync.RWMutex
	Count   int
	LogFile *os.File
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
	remoteList.mu.Lock()

	if _, exists := remoteList.Lists[args.List_ID]; !exists {
		remoteList.Lists[args.List_ID] = &List{
			List: make([]int, 0),
			Size: 0,
		}
		remoteList.Count++
		fmt.Printf("Created new list with ID %d\n", args.List_ID)
	}
	remoteList.mu.Unlock()

	list := remoteList.Lists[args.List_ID]
	list.mu.Lock()
	defer list.mu.Unlock()

	list.List = append(list.List, args.Value)
	list.Size++

	if remoteList.LogFile != nil {
		_, err := remoteList.LogFile.WriteString(fmt.Sprintf("APPEND: %d %d\n", args.List_ID, args.Value))
		if err != nil {
			fmt.Println("Error writing to log file:", err)
		}
	}

	*reply = true
	fmt.Printf("Lista[%d]: %v\n", args.List_ID, list.List)
	return nil
}

func (remoteList *RemoteList) Get(args GetArgs, reply *int) error {
	remoteList.mu.RLock()
	list, exists := remoteList.Lists[args.List_ID]
	remoteList.mu.RUnlock()

	if !exists {
		return nil
	}

	list.mu.Lock()
	defer list.mu.Unlock()

	if args.Index < 0 || args.Index >= len(list.List) {
		return nil
	}

	*reply = list.List[args.Index]
	return nil
}

func (remoteList *RemoteList) Remove(list_id int, reply *int) error {
	remoteList.mu.RLock()
	list, exists := remoteList.Lists[list_id]
	remoteList.mu.RUnlock()

	if !exists || len(list.List) == 0 {
		return nil
	}

	list.mu.Lock()
	defer list.mu.Unlock()

	*reply = list.List[len(list.List)-1]
	list.List = list.List[:len(list.List)-1]
	list.Size--

	if remoteList.LogFile != nil {
		_, err := remoteList.LogFile.WriteString(fmt.Sprintf("REMOVE: %d\n", list_id))
		if err != nil {
			fmt.Println("Error writing to log file:", err)
		}
	}

	fmt.Printf("Lista[%d]: %v\n", list_id, list.List)
	return nil
}

func (remoteList *RemoteList) Size(list_id int, reply *int) error {
	remoteList.mu.RLock()
	list, exists := remoteList.Lists[list_id]
	remoteList.mu.RUnlock()

	if !exists {
		*reply = 0
		return nil
	}

	list.mu.Lock()
	defer list.mu.Unlock()

	*reply = list.Size
	return nil
}

func NewRemoteList() *RemoteList {
	remoteList := &RemoteList{
		Lists: make(map[int]*List),
	}

	snapshotRestored := restoreFromLatestSnapshot(remoteList)

	if !snapshotRestored {
		_, err := os.Stat("logs.txt")
		fileExists := !os.IsNotExist(err)

		logFile, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Println("Error opening log file:", err)
			return remoteList
		}

		remoteList.LogFile = logFile

		if fileExists {
			fmt.Println("Log file found. Restoring previous state...")
			restoreFromLogFile(remoteList, logFile)
		} else {
			fmt.Println("No previous log file found. Starting with empty map.")
		}
	} else {
		logFile, err := os.OpenFile("logs.txt", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Println("Error opening log file:", err)
		} else {
			remoteList.LogFile = logFile
		}
	}

	go startSnapshotWorker(remoteList)

	return remoteList
}
