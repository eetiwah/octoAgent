package main

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/hpcloud/tail"
)

// Position holds X/Y/Z coordinates
type Position struct {
	X, Y, Z, E float64
}

// PositionTracker monitors serial.log for position updates
type PositionTracker struct {
	pos   Position
	mutex sync.RWMutex
	re    *regexp.Regexp
}

// NewPositionTracker initializes the tracker and starts monitoring
func NewPositionTracker(logPath string) (*PositionTracker, error) {
	tracker := &PositionTracker{
		re: regexp.MustCompile(`Send: N\d+ G1 X([\d.-]+)\s+Y([\d.-]+)(?:\s+Z([\d.-]+))?(?:\s+E([\d.-]+))?`),
	}

	// Start background monitoring
	go tracker.monitorLog(logPath)
	return tracker, nil
}

// monitorLog tails serial.log and updates position
func (t *PositionTracker) monitorLog(logPath string) {
	tailConfig := tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		Poll:      true, // Use polling for high-frequency writes
		Logger:    tail.DefaultLogger,
	}

	tailer, err := tail.TailFile(logPath, tailConfig)
	if err != nil {
		fmt.Printf("Error tailing serial.log: %v\n", err)
		return
	}
	defer tailer.Cleanup()

	for line := range tailer.Lines {
		if line.Err != nil {
			fmt.Printf("Error reading line: %v\n", line.Err)
			continue
		}
		t.parseLine(line.Text)
	}
}

// parseLine extracts position from a log line
func (t *PositionTracker) parseLine(line string) {
	matches := t.re.FindStringSubmatch(line)
	if len(matches) < 3 {
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Update X, Y (always present)
	if x, err := strconv.ParseFloat(matches[1], 64); err == nil {
		t.pos.X = x
	}
	if y, err := strconv.ParseFloat(matches[2], 64); err == nil {
		t.pos.Y = y
	}

	// Check for Z and E based on match length
	switch len(matches) {
	case 3:
		// Only X, Y; preserve Z, E
		//fmt.Printf("Updated position: X:%.2f Y:%.2f Z:%.2f E:%.2f\n", t.pos.X, t.pos.Y, t.pos.Z, t.pos.E)
	case 4:
		// X, Y, Z; preserve E
		if matches[3] != "" {
			if z, err := strconv.ParseFloat(matches[3], 64); err == nil {
				t.pos.Z = z
			}
		}
		//fmt.Printf("Updated position: X:%.2f Y:%.2f Z:%.2f E:%.2f\n", t.pos.X, t.pos.Y, t.pos.Z, t.pos.E)
	case 5:
		// X, Y, Z, E
		if z, err := strconv.ParseFloat(matches[3], 64); err == nil {
			t.pos.Z = z
		}
		if e, err := strconv.ParseFloat(matches[4], 64); err == nil {
			t.pos.E = e
		}
		//fmt.Printf("Updated position: X:%.2f Y:%.2f Z:%.2f E:%.2f\n", t.pos.X, t.pos.Y, t.pos.Z, t.pos.E)
	default:
		fmt.Printf("Error: unexpected matches length %v\n", len(matches))
	}
}

// GetPosition returns the current position
func (t *PositionTracker) GetPosition() Position {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.pos
}
