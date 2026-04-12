package scanner

import (
	"context"
	"fmt"
	"io"
	"log"
	"redscout/lib"
	"redscout/models"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *Scanner) FetchSlowLog() error {
	s.muRedis.Lock()
	defer s.muRedis.Unlock()

	slowLog, err := s.redis.SlowLogGet(s.ctx, s.Config.TopK).Result()

	models.SlowLogList(slowLog).Sort("Timestamp")
	s.State.SlowLogs = slowLog
	s.State.Updates <- s.State

	return err
}

func (s *Scanner) FetchRedisInfo() error {
	s.muRedis.Lock()
	defer s.muRedis.Unlock()

	infoStr, err := s.redis.Info(s.ctx).Result()
	if err != nil {
		return err
	}

	parsed := models.ParseInfo(infoStr)

	dHits := parsed.Stats.KeyspaceHits - s.State.RedisInfo.Stats.KeyspaceHits
	dMisses := parsed.Stats.KeyspaceMisses - s.State.RedisInfo.Stats.KeyspaceMisses

	if (dHits + dMisses) > 0 {
		parsed.Computed.HitRate = float64(dHits) / float64(dHits+dMisses)
	} else {
		parsed.Computed.HitRate = s.State.RedisInfo.Computed.HitRate
	}

	if !s.State.LastInfoCheck.IsZero() {
		currCPUTime := parsed.CPU.SystemTime + parsed.CPU.UserTime
		prevCPUTime := s.State.RedisInfo.CPU.UserTime + s.State.RedisInfo.CPU.SystemTime
		parsed.Computed.CPUUsage = (currCPUTime - prevCPUTime) * 1000 / float64(time.Now().UnixMilli()-s.State.LastInfoCheck.UnixMilli())
	}

	s.State.LastInfoCheck = time.Now()
	s.State.RedisInfo = &parsed

	s.State.Updates <- s.State

	return nil
}

func (s *Scanner) scanKeys() ([]string, error) {
	//Batch size for scanning keys
	scanSize := int64(lib.ScanBatchSize)

	var (
		collected []string
		scanned   int64
	)

	for {
		res, next, err := s.redis.Scan(s.ctx, s.State.Cursor, "*", scanSize).Result()
		if err != nil {
			log.Printf("scan error: %v", err)
			return nil, err
		}

		collected = append(collected, res...)
		scanned += int64(len(res))
		s.State.Cursor = next
		s.State.ScannedKeys += int64(len(res))
		s.State.ScanProgress = min(float64(scanned)/float64(s.Config.KeysScanSize)*50, 50)
		s.State.Updates <- s.State
		if next == 0 || scanned >= s.Config.KeysScanSize {
			break
		}
	}
	return collected, nil
}

type trip struct {
	key     string
	mem     *redis.IntCmd
	ttl     *redis.DurationCmd
	typeCmd *redis.StatusCmd
}

func (s *Scanner) ScanMemory() error {
	s.updateStatus("Scanning memory")

	s.muRedis.Lock()
	defer s.muRedis.Unlock()

	log.Printf("Memory scan started")

	s.State.TotalKeysToScan = s.Config.KeysScanSize
	s.updateStatus("Collecting keys")

	keys, err := s.scanKeys()
	if err != nil {
		return err
	}

	s.State.TotalKeysToScan = int64(len(keys))
	s.updateStatus("Scanning memory")

	if _, err := s.scanFile.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek scan file: %w", err)
	}

	processedKeys := int64(0)
	for i := 0; i < len(keys); i += lib.MemoryPipeBatchSize {
		pipe := s.redis.Pipeline()

		keyBatch := keys[i:min(i+lib.MemoryPipeBatchSize, len(keys))]

		var trips []trip
		for _, key := range keyBatch {
			tr := trip{key: key}
			tr.mem = pipe.MemoryUsage(s.ctx, key)
			tr.ttl = pipe.TTL(s.ctx, key)
			tr.typeCmd = pipe.Type(s.ctx, key)
			trips = append(trips, tr)
		}

		_, _ = pipe.Exec(s.ctx)

		for _, tr := range trips {
			xMem, e1 := tr.mem.Result()
			xTtl, e2 := tr.ttl.Result()
			xType, e3 := tr.typeCmd.Result()
			if e1 != nil || e2 != nil || e3 != nil {
				continue
			}
			_, _ = s.scanFile.WriteString(fmt.Sprintf("%s %d %d %s\n", tr.key, xMem, int64(xTtl.Seconds()), xType))
		}

		processedKeys += int64(len(keyBatch))
		s.State.ScanProgress = 50 + min(float64(processedKeys)/float64(len(keys))*50, 49)
		s.State.Updates <- s.State
	}

	s.State.ScanProgress = 100
	log.Printf("Memory scan completed; scanned %d keys", len(keys))
	s.updateStatus("Memory scan completed")
	return nil
}

func (s *Scanner) MonitorOps() error {
	s.muRedis.Lock()
	defer s.muRedis.Unlock()
	s.updateStatus("Monitoring operations")

	log.Printf("Ops monitor started for %v", s.Config.MonitorDuration)

	s.State.MonitorStartTime = time.Now()
	s.State.MonitorDurationTotal = s.Config.MonitorDuration
	s.State.MonitorProgress = 0

	//Buffered channel to handle Redis monitor output
	ch := make(chan string, 10000)
	defer close(ch)

	client, err := lib.RedisClientFromConfig(s.Config)
	if err != nil {
		log.Printf("Error creating Redis redis for monitoring: %v", err)
		return err
	}
	defer client.Close()

	ctxTimeout, cancel := context.WithTimeout(s.ctx, s.Config.MonitorDuration)
	defer cancel()

	monitor := client.Monitor(ctxTimeout, ch)
	monitor.Start()
	defer monitor.Stop()

	if _, err := s.monitorFile.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek monitor file: %w", err)
	}

	progressTicker := time.NewTicker(100 * time.Millisecond)
	defer progressTicker.Stop()

	for {
		select {
		case line, ok := <-ch:
			if !ok {
				continue
			}
			parts := strings.Split(line, "\"")
			if len(parts) < 2 {
				continue
			}
			cmd := strings.ToLower(parts[1])
			if cmd == "eval" {
				continue
			}

			// Extract keys from quoted args (parts[3], parts[5], parts[7], ...)
			var keys []string
			switch cmd {
			case "mset", "msetnx":
				// MSET key1 val1 key2 val2 ... — every other arg is a key
				for j := 3; j < len(parts); j += 4 {
					keys = append(keys, parts[j])
				}
			case "mget", "del", "unlink", "exists", "touch", "watch":
				// All args are keys
				for j := 3; j < len(parts); j += 2 {
					keys = append(keys, parts[j])
				}
			default:
				if len(parts) >= 4 {
					keys = append(keys, parts[3])
				}
			}

			for _, key := range keys {
				_, _ = s.monitorFile.WriteString(fmt.Sprintf("%s %s\n", key, cmd))
			}
		case <-progressTicker.C:
			elapsed := time.Since(s.State.MonitorStartTime)
			s.State.MonitorProgress = min(float64(elapsed)/float64(s.Config.MonitorDuration)*100, 100)
			s.State.Updates <- s.State
		case <-ctxTimeout.Done():
			s.State.MonitorProgress = 100
			s.State.TotalMonitorDuration += s.Config.MonitorDuration
			s.State.Updates <- s.State
			s.updateStatus("Monitoring completed")
			log.Printf("Monitoring completed")
			return nil
		}
	}
}

func (s *Scanner) InfoUpdates() {
	ticker := time.NewTicker(s.Config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			err := s.FetchRedisInfo()
			if err != nil {
				log.Printf("Error fetching Redis info: %v", err)
				continue
			}
		}
	}
}
