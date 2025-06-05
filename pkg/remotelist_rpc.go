package remoteList

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type List struct {
	mu   sync.Mutex
	List []int
	Size int
}

type RemoteList struct {
	Lists   []List
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
	if args.List_ID < 0 || args.List_ID >= remoteList.Count {
		*reply = false
		return nil
	}

	remoteList.Lists[args.List_ID].mu.Lock()
	defer remoteList.Lists[args.List_ID].mu.Unlock()

	remoteList.Lists[args.List_ID].List = append(remoteList.Lists[args.List_ID].List, args.Value)
	remoteList.Lists[args.List_ID].Size++

	if remoteList.LogFile != nil {
		_, err := remoteList.LogFile.WriteString(fmt.Sprintf("APPEND: %d %d\n", args.List_ID, args.Value))

		if err != nil {
			fmt.Println("Error writing to log file:", err)
		}
	}

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

	if remoteList.LogFile != nil {
		_, err := remoteList.LogFile.WriteString(fmt.Sprintf("REMOVE: %d\n", list_id))

		if err != nil {
			fmt.Println("Error writing to log file:", err)
		}
	}

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
	remoteList := new(RemoteList)
	remoteList.Lists = make([]List, 3)
	remoteList.Count = 3

	for i := 0; i < 3; i++ {
		remoteList.Lists[i].List = make([]int, 0)
		remoteList.Lists[i].Size = 0
		fmt.Printf("List %d initialized.\n", i)
	}

	_, err := os.Stat("logs.txt")
	fileExists := !os.IsNotExist(err)

	logFile, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		fmt.Println("Error opening log file:", err)
		return nil
	}

	remoteList.LogFile = logFile

	if fileExists {
		fmt.Println("Log file found. Restoring previous state...")

		logFile.Seek(0, 0)

		scanner := bufio.NewScanner(logFile)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "APPEND:") {
				var listID, value int
				fmt.Sscanf(line, "APPEND: %d %d", &listID, &value)

				if listID >= 0 && listID < remoteList.Count {
					remoteList.Lists[listID].List = append(remoteList.Lists[listID].List, value)
					remoteList.Lists[listID].Size++
					fmt.Printf("Restored: APPEND %d to list %d\n", value, listID)
				}
			} else if strings.HasPrefix(line, "REMOVE:") {
				var listID int
				fmt.Sscanf(line, "REMOVE: %d", &listID)

				if listID >= 0 && listID < remoteList.Count && len(remoteList.Lists[listID].List) > 0 {
					remoteList.Lists[listID].List = remoteList.Lists[listID].List[:len(remoteList.Lists[listID].List)-1]
					remoteList.Lists[listID].Size--
					fmt.Printf("Restored: REMOVE from list %d\n", listID)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading log file:", err)
		}

		for i := 0; i < remoteList.Count; i++ {
			fmt.Printf("List %d restored: %v\n", i, remoteList.Lists[i].List)
		}
	} else {
		fmt.Println("No previous log file found. Starting with empty lists.")
	}

	return remoteList
}
