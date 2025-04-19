package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/hpcloud/tail"
)

// Position holds X/Y/Z coordinates
type Position struct {
	X, Y, Z, E float64
}

// PositionTracker monitors serial.log for position updates
type PositionTracker struct {
	pos        Position
	mutex      sync.RWMutex
	re         *regexp.Regexp
	reZ        *regexp.Regexp   // For G1 Z without X/Y
	zMutex     sync.RWMutex     // Mutex for zValues
	zValues    map[float64]bool // Unique Z values
	storeZ     bool             // Flag to store Z values
	outputFile string           // Output file for Z values
}

// NewPositionTracker initializes the tracker and starts monitoring
func NewPositionTracker(logPath string, storeZ bool, outputFile string) (*PositionTracker, error) {
	tracker := &PositionTracker{
		re:         regexp.MustCompile(`Send: N\d+ G[0-1] X([-]?[0-9]+(?:\.[0-9]+)?)\s+Y([-]?[0-9]+(?:\.[0-9]+)?)(?:\s+Z([-]?[0-9]+(?:\.[0-9]+)?))?(?:\s+E([-]?[0-9]+(?:\.[0-9]+)?))?`),
		reZ:        regexp.MustCompile(`Send: N\d+ G[0-1](?:\s+F\d+)?\s+Z([-]?[0-9]+(?:\.[0-9]+)?)`),
		zValues:    make(map[float64]bool),
		storeZ:     storeZ,
		outputFile: outputFile,
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
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Try regex with X/Y
	matches := t.re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		// Update X, Y
		if x, err := strconv.ParseFloat(matches[1], 64); err == nil {
			t.pos.X = x
		}
		if y, err := strconv.ParseFloat(matches[2], 64); err == nil {
			t.pos.Y = y
		}

		// Check for Z and E
		if len(matches) >= 4 && matches[3] != "" {
			if z, err := strconv.ParseFloat(matches[3], 64); err == nil {
				t.pos.Z = z
				if t.storeZ {
					t.zMutex.Lock()
					t.zValues[z] = true
					t.zMutex.Unlock()
					t.saveZValues()
				}
				if strings.Contains(line, "G1") {
					log.Printf("Printing Z position (G1): %.3f", z)
				} else {
					log.Printf("Travel Z position (G0): %.3f", z)
				}
			}
		}
		if len(matches) >= 5 && matches[4] != "" {
			if e, err := strconv.ParseFloat(matches[4], 64); err == nil {
				t.pos.E = e
			}
		}
		return
	}

	// Try regex for G1/G0 Z without X/Y
	matches = t.reZ.FindStringSubmatch(line)
	if len(matches) == 2 {
		if z, err := strconv.ParseFloat(matches[1], 64); err == nil {
			t.pos.Z = z
			if t.storeZ {
				t.zMutex.Lock()
				t.zValues[z] = true
				t.zMutex.Unlock()
				t.saveZValues()
			}
			if strings.Contains(line, "G1") {
				log.Printf("Printing Z position (G1, Z-only): %.3f", z)
			} else {
				log.Printf("Travel Z position (G0, Z-only): %.3f", z)
			}
		}
	}
}

// *** we are opening, reading the whole file, sorting, and adding then closing every time.

// saveZValues writes unique Z values to the output file
func (t *PositionTracker) saveZValues() {
	if !t.storeZ {
		return
	}
	t.zMutex.RLock()
	defer t.zMutex.RUnlock()

	zValues := make([]float64, 0, len(t.zValues))
	for z := range t.zValues {
		zValues = append(zValues, z)
	}

	sort.Float64s(zValues)

	file, err := os.Create(t.outputFile)
	if err != nil {
		log.Printf("Failed to write %s: %v", t.outputFile, err)
		return
	}
	defer file.Close()

	for _, z := range zValues {
		fmt.Fprintf(file, "%.3f\n", z)
	}
	log.Printf("Saved %d unique Z values to %s", len(zValues), t.outputFile)
}

// GetPosition returns the current position
func (t *PositionTracker) GetPosition() Position {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.pos
}
