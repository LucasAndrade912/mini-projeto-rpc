package remoteList

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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
			fmt.Println("No previous log file found. Starting with empty lists.")
		}
	} else {
		logFile, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Println("Error opening log file:", err)
		} else {
			remoteList.LogFile = logFile
		}
	}

	go startSnapshotWorker(remoteList)

	return remoteList
}

func startSnapshotWorker(remoteList *RemoteList) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	fmt.Println("Snapshot worker started. Snapshots will be created every minute.")

	for range ticker.C {
		createSnapshot(remoteList)
	}
}

func createSnapshot(remoteList *RemoteList) {
	for i := 0; i < remoteList.Count; i++ {
		remoteList.Lists[i].mu.Lock()
	}

	defer func() {
		for i := 0; i < remoteList.Count; i++ {
			remoteList.Lists[i].mu.Unlock()
		}
	}()

	snapshotsDir := "snapshots"
	if _, err := os.Stat(snapshotsDir); os.IsNotExist(err) {
		err := os.Mkdir(snapshotsDir, 0755)
		if err != nil {
			fmt.Printf("Error creating snapshots directory: %v\n", err)
			return
		}
		fmt.Println("Created snapshots directory")
	}

	timestamp := time.Now().Format("20060102_150405")
	snapshotFilename := fmt.Sprintf("%s/snapshot_%s.txt", snapshotsDir, timestamp)

	snapshotFile, err := os.Create(snapshotFilename)
	if err != nil {
		fmt.Printf("Error creating snapshot file: %v\n", err)
		return
	}
	defer snapshotFile.Close()

	snapshotFile.WriteString(fmt.Sprintf("SNAPSHOT CREATED AT: %s\n", timestamp))
	snapshotFile.WriteString("----------------------------------------\n")

	for i := 0; i < remoteList.Count; i++ {
		snapshotFile.WriteString(fmt.Sprintf("LIST %d (size=%d): ", i, remoteList.Lists[i].Size))

		if len(remoteList.Lists[i].List) > 0 {
			for j, value := range remoteList.Lists[i].List {
				if j > 0 {
					snapshotFile.WriteString(", ")
				}
				snapshotFile.WriteString(fmt.Sprintf("%d", value))
			}
		} else {
			snapshotFile.WriteString("empty")
		}

		snapshotFile.WriteString("\n")
	}

	fmt.Printf("Snapshot created: %s\n", snapshotFilename)
}

func restoreFromLogFile(remoteList *RemoteList, logFile *os.File) {
	_, err := logFile.Seek(0, io.SeekCurrent)
	if err != nil {
		fmt.Println("Error getting current file position:", err)
	}

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
				fmt.Printf("Restored from logs: APPEND %d to list %d\n", value, listID)
			}
		} else if strings.HasPrefix(line, "REMOVE:") {
			var listID int
			fmt.Sscanf(line, "REMOVE: %d", &listID)

			if listID >= 0 && listID < remoteList.Count && len(remoteList.Lists[listID].List) > 0 {
				remoteList.Lists[listID].List = remoteList.Lists[listID].List[:len(remoteList.Lists[listID].List)-1]
				remoteList.Lists[listID].Size--
				fmt.Printf("Restored from logs: REMOVE from list %d\n", listID)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading log file:", err)
	}

	logFile.Seek(0, io.SeekEnd)

	for i := 0; i < remoteList.Count; i++ {
		fmt.Printf("List %d restored from logs: %v\n", i, remoteList.Lists[i].List)
	}
}

func restoreFromLatestSnapshot(remoteList *RemoteList) bool {
	snapshotsDir := "snapshots"
	if _, err := os.Stat(snapshotsDir); os.IsNotExist(err) {
		return false
	}

	files, err := os.ReadDir(snapshotsDir)
	if err != nil {
		fmt.Printf("Error reading snapshots directory: %v\n", err)
		return false
	}

	if len(files) == 0 {
		return false
	}

	var latestSnapshot string
	var latestTime time.Time
	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), "snapshot_") {
			continue
		}

		timeStr := strings.TrimPrefix(file.Name(), "snapshot_")
		timeStr = strings.TrimSuffix(timeStr, ".txt")
		snapTime, err := time.Parse("20060102_150405", timeStr)

		if err == nil && (latestSnapshot == "" || snapTime.After(latestTime)) {
			latestSnapshot = file.Name()
			latestTime = snapTime
		}
	}

	if latestSnapshot == "" {
		return false
	}

	snapshotPath := filepath.Join(snapshotsDir, latestSnapshot)
	file, err := os.Open(snapshotPath)
	if err != nil {
		fmt.Printf("Error opening latest snapshot: %v\n", err)
		return false
	}
	defer file.Close()

	fmt.Printf("Restoring from latest snapshot: %s\n", latestSnapshot)

	scanner := bufio.NewScanner(file)
	scanner.Scan() // SNAPSHOT CREATED AT: ...
	scanner.Scan() // ------------------------

	listIndex := 0
	for scanner.Scan() && listIndex < remoteList.Count {
		line := scanner.Text()

		if !strings.HasPrefix(line, "LIST ") {
			continue
		}

		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue
		}

		valuesStr := strings.TrimSpace(line[colonIndex+1:])
		if valuesStr == "empty" {
			listIndex++
			continue
		}

		valueStrs := strings.Split(valuesStr, ", ")
		for _, valueStr := range valueStrs {
			var value int
			if _, err := fmt.Sscanf(valueStr, "%d", &value); err == nil {
				remoteList.Lists[listIndex].List = append(remoteList.Lists[listIndex].List, value)
				remoteList.Lists[listIndex].Size++
			}
		}

		listIndex++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading snapshot file: %v\n", err)
		return false
	}

	for i := 0; i < remoteList.Count; i++ {
		fmt.Printf("List %d restored from snapshot: %v\n", i, remoteList.Lists[i].List)
	}

	return true
}
