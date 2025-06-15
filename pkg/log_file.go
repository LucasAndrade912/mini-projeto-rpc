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
