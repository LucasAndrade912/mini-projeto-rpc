package remoteList

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func startSnapshotWorker(remoteList *RemoteList) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	fmt.Println("Snapshot worker started. Snapshots will be created every minute.")

	for range ticker.C {
		createSnapshot(remoteList)
	}
}

func createSnapshot(remoteList *RemoteList) {
	remoteList.mu.RLock()

	for _, list := range remoteList.Lists {
		list.mu.Lock()
	}

	defer func() {
		for _, list := range remoteList.Lists {
			list.mu.Unlock()
		}
		remoteList.mu.RUnlock()
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

	var listIDs []int
	for listID := range remoteList.Lists {
		listIDs = append(listIDs, listID)
	}
	sort.Ints(listIDs)

	for _, listID := range listIDs {
		list := remoteList.Lists[listID]
		snapshotFile.WriteString(fmt.Sprintf("LIST %d (size=%d): ", listID, list.Size))

		if len(list.List) > 0 {
			for j, value := range list.List {
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

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "LIST ") {
			continue
		}

		var listID int
		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue
		}

		listPart := line[:colonIndex]
		fmt.Sscanf(listPart, "LIST %d", &listID)

		valuesStr := strings.TrimSpace(line[colonIndex+1:])

		if _, exists := remoteList.Lists[listID]; !exists {
			remoteList.Lists[listID] = &List{
				List: make([]int, 0),
				Size: 0,
			}
		}

		if valuesStr == "empty" {
			continue
		}

		valueStrs := strings.Split(valuesStr, ", ")
		for _, valueStr := range valueStrs {
			var value int
			if _, err := fmt.Sscanf(valueStr, "%d", &value); err == nil {
				remoteList.Lists[listID].List = append(remoteList.Lists[listID].List, value)
				remoteList.Lists[listID].Size++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading snapshot file: %v\n", err)
		return false
	}

	for listID, list := range remoteList.Lists {
		fmt.Printf("List %d restored from snapshot: %v\n", listID, list.List)
	}

	return true
}
