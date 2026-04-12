package scanner

import (
	"bufio"
	"container/heap"
	"io"
	"log"
	"redscout/models"
	"strconv"
	"strings"
)

func (s *Scanner) ComputeNamespaceStats() error {
	log.Printf(
		"Generating namespace stats for prefix: %s",
		strings.Join(s.State.CurrentPrefix, s.Config.Delimiter),
	)
	snapshots := make(map[string]*models.NamespaceSnapshot)

	if err := s.computeScanOpsLogs(snapshots); err != nil {
		return err
	}

	if err := s.computeNamespaceMonitorLog(snapshots); err != nil {
		return err
	}

	var metrics models.NamespaceMetricList
	for _, snapshot := range snapshots {
		metrics = append(metrics, snapshot.ToMetric(s.State))
	}

	if len(metrics) == 0 {
		log.Printf(
			"No namespace metrics found for prefix: %s",
			strings.Join(s.State.CurrentPrefix, s.Config.Delimiter),
		)
	}

	s.State.NamespaceStats = metrics
	s.State.NamespaceStats.Sort("Keys")
	s.State.Updates <- s.State

	return nil
}

func (s *Scanner) computeScanOpsLogs(snapshots map[string]*models.NamespaceSnapshot) error {
	s.muScan.Lock()
	defer s.muScan.Unlock()

	log.Printf(
		"Processing scan log for prefix: %s\n",
		strings.Join(s.State.CurrentPrefix, s.Config.Delimiter),
	)

	_, err := s.scanFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(s.scanFile)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 4 {
			continue
		}

		key := s.kp.NewKey(parts[0], false)
		namespace, err := s.kp.Namespace(key, s.State.CurrentPrefix, true)
		if err != nil {
			continue
		}

		memory, _ := strconv.ParseInt(parts[1], 10, 64)
		ttl, _ := strconv.ParseInt(parts[2], 10, 64)
		keyType := parts[3]

		snapshot, exists := snapshots[namespace]
		if !exists {
			snapshot = &models.NamespaceSnapshot{
				Namespace:    namespace,
				OpsFrequency: make(map[string]int64),
				Types:        make([]string, 0),
			}
			snapshots[namespace] = snapshot
		}

		snapshot.Keys++
		snapshot.TotalMemory += memory
		if ttl > 0 {
			snapshot.KeysWithTTL++
			snapshot.TotalTTL += ttl
		}

		typeExists := false
		for _, t := range snapshot.Types {
			if t == keyType {
				typeExists = true
				break
			}
		}
		if !typeExists {
			snapshot.Types = append(snapshot.Types, keyType)
		}
	}

	return scanner.Err()
}

func (s *Scanner) computeNamespaceMonitorLog(
	snapshots map[string]*models.NamespaceSnapshot,
) error {
	s.muMonitor.Lock()
	defer s.muMonitor.Unlock()

	log.Printf("Processing monitor log for prefix: %s\n", strings.Join(s.State.CurrentPrefix, s.Config.Delimiter))
	_, err := s.monitorFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(s.monitorFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		key := s.kp.NewKey(parts[0], false)
		namespace, err := s.kp.Namespace(key, s.State.CurrentPrefix, true)
		if err != nil {
			continue
		}

		command := parts[1]

		snapshot, exists := snapshots[namespace]
		if !exists {
			snapshot = &models.NamespaceSnapshot{
				Namespace:    namespace,
				OpsFrequency: make(map[string]int64),
				Types:        make([]string, 0),
			}
			snapshots[namespace] = snapshot
		}

		snapshot.OpsFrequency[command]++
	}

	return scanner.Err()
}

// ComputeBigKeysFromScanLog returns the top n keys by memory usage from the scan log using a min-heap.
func (s *Scanner) ComputeBigKeysFromScanLog() error {
	s.muScan.Lock()
	defer s.muScan.Unlock()

	_, err := s.scanFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(s.scanFile)
	h := &models.BigKeyMinHeap{}
	heap.Init(h)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 4 {
			continue
		}
		keyStr := parts[0]
		key := s.kp.NewKey(keyStr, false)

		memory, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		ttl, _ := strconv.ParseInt(parts[2], 10, 64)
		keyType := parts[3]
		bk := models.BigKey{Key: key, Size: memory, Type: keyType, TTL: ttl}
		if int64(h.Len()) < s.Config.TopK {
			heap.Push(h, bk)
		} else if h.Len() > 0 && (*h)[0].Size < memory {
			heap.Pop(h)
			heap.Push(h, bk)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	// Extract from heap to slice, largest first
	result := make(models.BigKeyList, h.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(models.BigKey)
	}
	s.State.BigKeys = result
	s.State.Updates <- s.State
	return nil
}

func (s *Scanner) ComputeHotKeysFromMonitorLog() error {
	s.muMonitor.Lock()
	defer s.muMonitor.Unlock()

	_, err := s.monitorFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(s.monitorFile)

	type keyStats struct {
		ops      int64
		commands map[string]int64
	}

	keyData := make(map[string]*keyStats)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		keyStr := parts[0]
		cmd := parts[1]

		stats, exists := keyData[keyStr]
		if !exists {
			stats = &keyStats{commands: make(map[string]int64)}
			keyData[keyStr] = stats
		}
		stats.ops++
		stats.commands[cmd]++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	duration := s.State.TotalMonitorDuration.Seconds()
	if duration == 0 {
		duration = 1
	}

	h := &models.HotKeyMinHeap{}
	heap.Init(h)
	for k, stats := range keyData {
		opsPerSec := float64(stats.ops) / duration

		topCmd := ""
		topCmdCount := int64(0)
		for cmd, count := range stats.commands {
			if count > topCmdCount {
				topCmd = cmd
				topCmdCount = count
			}
		}

		hk := models.HotKey{
			Key:     s.kp.NewKey(k, false),
			Ops:     opsPerSec,
			Command: strings.ToUpper(topCmd),
		}
		if int64(h.Len()) < s.Config.TopK {
			heap.Push(h, hk)
		} else if h.Len() > 0 && (*h)[0].Ops < opsPerSec {
			heap.Pop(h)
			heap.Push(h, hk)
		}
	}
	result := make(models.HotKeyList, h.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(models.HotKey)
	}
	s.State.HotKeys = result
	s.State.Updates <- s.State
	return nil
}
