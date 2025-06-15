package remoteList

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
