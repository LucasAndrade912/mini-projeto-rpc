package remoteList

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

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

			if _, exists := remoteList.Lists[listID]; !exists {
				remoteList.Lists[listID] = &List{
					List: make([]int, 0),
					Size: 0,
				}
			}

			remoteList.Lists[listID].List = append(remoteList.Lists[listID].List, value)
			remoteList.Lists[listID].Size++
			fmt.Printf("Restored from logs: APPEND %d to list %d\n", value, listID)

		} else if strings.HasPrefix(line, "REMOVE:") {
			var listID int
			fmt.Sscanf(line, "REMOVE: %d", &listID)

			if list, exists := remoteList.Lists[listID]; exists && len(list.List) > 0 {
				list.List = list.List[:len(list.List)-1]
				list.Size--
				fmt.Printf("Restored from logs: REMOVE from list %d\n", listID)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading log file:", err)
	}

	logFile.Seek(0, io.SeekEnd)

	for listID, list := range remoteList.Lists {
		fmt.Printf("List %d restored from logs: %v\n", listID, list.List)
	}
}

func clearLogFile(remoteList *RemoteList) {
	if remoteList.LogFile == nil {
		return
	}

	remoteList.LogFile.Close()

	logFile, err := os.OpenFile("logs.txt", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error clearing log file: %v\n", err)
		return
	}

	remoteList.LogFile = logFile
	fmt.Println("Log file cleared after snapshot creation.")
}
