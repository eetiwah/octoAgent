package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"octoAgent/octoprint"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var isConnected = false

type Job struct {
	File struct {
		Name     string      `json:"name"`
		Path     string      `json:"path"`
		Type     string      `json:"type"`
		TypePath interface{} `json:"typePath"`
		Hash     string      `json:"hash"`
		Size     int         `json:"size"`
		Date     int64       `json:"date"`
		Origin   string      `json:"origin"`
		Refs     struct {
			Resource string `json:"resource"`
			Download string `json:"download"`
			Model    string `json:"model"`
		} `json:"refs"`
		GcodeAnalysis struct {
			EstimatedPrintTime int `json:"estimatedPrintTime"`
			Filament           struct {
				Length int `json:"length"`
				Volume int `json:"volume"`
			} `json:"filament"`
		} `json:"gcodeAnalysis"`
		Print struct {
			Failure int64 `json:"failure"`
			Success int64 `json:"success"`
			Last    struct {
				Date    int64 `json:"date"`
				Success bool  `json:"success"`
			} `json:"last"`
		} `json:"print"`
	} `json:"file"`
	EstimatedPrintTime int `json:"estimatedPrintTime"`
	LastPrintTime      int `json:"lastPrintTime"`
	Filament           struct {
		Length int `json:"length"`
		Volume int `json:"volume"`
	} `json:"filament"`
	Filepos int `json:"filepos"`
}

type JobStatus struct {
	FileName    string  `json:"file_name"`
	Progress    float64 `json:"progress"`
	TimeElapsed float64 `json:"time_elapsed"`
	TimeLeft    float64 `json:"time_left"`
}

type Progress struct {
	Completion    float64 `json:"completion"`
	Filepos       int     `json:"filepos"`
	PrintTime     int     `json:"printTime"`
	PrintTimeLeft int     `json:"printTimeLeft"`
}

type FileInfo struct {
	Name          string                             `json:"name"`
	Path          string                             `json:"path"`
	Type          string                             `json:"type"`
	TypePath      []string                           `json:"typePath"`
	Hash          string                             `json:"hash"`
	Size          uint64                             `json:"size"`
	Date          octoprint.JSONTime                 `json:"date"`
	Origin        string                             `json:"origin"`
	Refs          octoprint.Reference                `json:"refs"`
	GcodeAnalysis octoprint.GCodeAnalysisInformation `json:"gcodeAnalysis"`
	Print         octoprint.PrintStats               `json:"print"`
}

type Refs struct {
	Resource string `json:"resource"`
	Download string `json:"download"`
	Model    string `json:"model"`
}

type GcodeAnalysis struct {
	EstimatedPrintTime float64 `json:"estimatedPrintTime"`
	Filament           struct {
		Length uint32  `json:"length"`
		Volume float64 `json:"volume"`
	} `json:"filament"`
}

type Filament struct {
	Length float64 `json:"length"`
	Volume float64 `json:"volume"`
}

type Print struct {
	Failure int64 `json:"failure"`
	Success int64 `json:"success"`
	Last    struct {
		Date    int64 `json:"date"`
		Success bool  `json:"success"`
	} `json:"last"`
}

type PrinterStatus struct {
	State        string  `json:"state"`
	ExtruderTemp float64 `json:"extruder_temp"`
	BedTemp      float64 `json:"bed_temp"`
	Progress     float64 `json:"progress"`
	FileName     string  `json:"file_name"`
	TimeLeft     float64 `json:"time_left"`
	Printing     bool    `json:"printing"`
}

type PrintStatus struct {
	FileName     string  `json:"file_name"`
	Progress     float64 `json:"progress"`
	TimeElapsed  float64 `json:"time_elapsed"`
	TimeLeft     float64 `json:"time_left"`
	State        string  `json:"state"`
	PositionX    float64 `json:"position_x"`
	PositionY    float64 `json:"position_y"`
	PositionZ    float64 `json:"position_z"`
	PositionE    float64 `json:"position_e"`
	ExtruderTemp float64 `json:"extruder_temp"`
	BedTemp      float64 `json:"bed_temp"`
}

type FilesResponse struct {
	Files []FileInfo `json:"Files"`
	Free  int        `json:"Free"`
}

type TempState struct {
	Tool   string
	Time   string
	Actual float64
	Target float64
}

type ConnectionSettings struct {
	IsConnecting bool
	IsError      bool
	IsOffline    bool
	IsOperation  bool
	IsPrinting   bool
	BaudRate     int    `json:"baudrate"`
	Port         string `json:"port"`
}

func GetApiVersion() string {
	if isConnected {
		octoReq := octoprint.VersionRequest{}
		s, err := octoReq.Do(octoclient)
		if err != nil {
			return ("Error: " + err.Error())
		}

		return s.API
	} else {
		return "Error: not connected"
	}
}

func CheckOctoService() string {
	if _checkOctoService() {
		return "OctoService is running"
	} else {
		return "OctoService is NOT running"
	}
}

func _checkOctoService() bool {
	cmd := exec.Command("pgrep", "-f", "octoprint serve")
	err := cmd.Run()
	return err == nil // Returns true if process is found
}

func StartOctoService() string {
	user, err := user.Current()
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return "Octoprint server was NOT started"
	}
	octoPath := path.Join(user.HomeDir, "Projects/octo/OctoPrint/venv/bin/octoprint")

	// Check if OctoPrint is already running
	if _checkOctoService() {
		return "Octoprint server is already running"
	}

	// Create command with serve argument
	cmd := exec.Command(octoPath, "serve")

	// Start as background process
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting OctoPrint: %v", err)
		return "Octoprint server was NOT started"
	}

	// Give it a moment to start, then verify
	time.Sleep(2 * time.Second)
	if _checkOctoService() {
		log.Println("Octoprint server was started")
		return "Octoprint server was started, now connect to it"
	}

	log.Printf("Octoprint started but not detected running")
	return "Octoprint server was NOT started"
}

func ConnectOctoService() string {
	// Check if OctoPrint is running
	if !_checkOctoService() {
		return "Octoprint server is NOT already running, please start it"
	}

	if !isConnected {
		octoReq := octoprint.ConnectRequest{
			//		Autoconnect: true,
		}
		err := octoReq.Do(octoclient)
		if err != nil {
			return ("Error: " + err.Error())
		}

		isConnected = true
		return "Connected"
	} else {
		return "Already connected"
	}
}

func DisconnectOctoService() string {
	if !_checkOctoService() {
		return "Octoprint server is NOT already running, disconnect is not needed"
	}

	if isConnected {
		octoReq := octoprint.DisconnectRequest{}
		err := octoReq.Do(octoclient)
		if err != nil {
			return ("Error: " + err.Error())
		}

		isConnected = false
		str := "Disconnected"
		return str
	} else {
		return "Already disconnected"
	}
}

func GetConnectionSettings() string {
	// Check if OctoPrint is running
	if !_checkOctoService() {
		return "Error: octoprint server is NOT already running"
	}

	if isConnected {
		octoReq := octoprint.ConnectionRequest{}
		settings, err := octoReq.Do(octoclient)
		if err != nil {
			return ("Error: " + err.Error())
		}

		connectionSettings := ConnectionSettings{
			IsConnecting: settings.Current.State.IsConnecting(),
			IsError:      settings.Current.State.IsError(),
			IsOffline:    settings.Current.State.IsOffline(),
			IsOperation:  settings.Current.State.IsOperational(),
			IsPrinting:   settings.Current.State.IsPrinting(),
			BaudRate:     settings.Current.BaudRate,
			Port:         settings.Current.Port,
		}
		// Encode to JSON
		jsonBytes, err := json.Marshal(connectionSettings)
		if err != nil {
			log.Printf("Error: failed to encode connection setttings: %v", err)
			return "Error: failed to encode connection settings: " + err.Error()
		}

		return string(jsonBytes)
	} else {
		return "Error: not connected"
	}
}

func GetOctoFileInfo(commandList []string) string {
	switch len(commandList) {
	case 2:
		if commandList[1] == "-help" {
			return `{"usage":"getfileinfo filename.ext"}`
		}

		if !isConnected {
			return `{"error":"not connected to octoService"}`
		}

		// Create file request
		octoReq := octoprint.FileRequest{
			Location:  octoprint.Local,
			Filename:  commandList[1], // e.g., "Ring.gcode"
			Recursive: false,
		}

		// Execute request
		response, err := octoReq.Do(octoclient)
		if err != nil {
			log.Printf("GetFileInfo error for %s: %v", commandList[1], err)
			return fmt.Sprintf(`{"error":"failed to get file info: %v"}`, err)
		}

		// Log raw response
		rawBytes, _ := json.Marshal(response)
		log.Printf("Raw file info for %s: %s", commandList[1], rawBytes)

		// Map to FileInfo
		fileInfo := FileInfo{
			Name:          response.Name,
			Path:          response.Path,
			Type:          response.Type,
			TypePath:      response.TypePath,
			Hash:          response.Hash,
			Size:          response.Size,
			Date:          response.Date,
			Origin:        response.Origin,
			Refs:          response.Refs,
			GcodeAnalysis: response.GCodeAnalysis,
			Print:         response.Print,
		}

		// Encode to JSON
		jsonBytes, err := json.Marshal(fileInfo)
		if err != nil {
			log.Printf("Error: encoding file info: %v", err)
			return `{"Error":"failed to encode file info"}`
		}

		return string(jsonBytes)

	default:
		return `{"Error":"parameter mismatch"}`
	}
}

func GetOctoFileList() string {
	if isConnected {
		octoReq := octoprint.FilesRequest{
			Location:  octoprint.Local,
			Recursive: false,
		}

		jsonResponse, err := octoReq.Do(octoclient)
		if err != nil {
			return ("Error GetFileList: " + err.Error())
		}

		// Convert jobResponse to JSON bytes
		jsonData, err := json.Marshal(jsonResponse)
		if err != nil {
			return "Error converting to JSON: " + err.Error()
		}

		// Create an instance of FilesResponse struct to hold the JSON data
		var filesResponse FilesResponse

		// Unmarshal the JSON data into the FilesResponse struct
		if err := json.Unmarshal(jsonData, &filesResponse); err != nil {
			return "Error unmarshalling JSON: " + err.Error()
		}

		// Build result
		var strB strings.Builder
		for _, file := range filesResponse.Files {
			fmt.Fprintf(&strB, "Name: %s\nSize: %d\nEstTime: %.1f\n", file.Name, file.Size, file.GcodeAnalysis.EstimatedPrintTime)
		}

		str := strB.String()
		return str

	} else {
		return "Error: not connected"
	}
}

/*
func AddFile(c *octoprint.Client, filename string, fileContent []byte) string {
	if isConnected {
		octoReq := octoprint.UploadFileRequest{
			Location: octoprint.Local,
			Select:   false,
			Print:    false,
		}

		// Add file to request
		err := octoReq.AddFile(filename, bytes.NewBuffer(fileContent))
		if err != nil {
			return ("Error adding file to request: " + err.Error())
		}

		// Perform the add
		response, err := octoReq.Do(c)
		if err != nil {
			return ("Error adding file: " + err.Error())
		}

		return response.File.Local.Name + " was added"

	} else {
		return "Error: not connected"
	}
}
*/

func DeleteOctoFile(commandList []string) string {
	filename := commandList[1]

	if isConnected {
		octoReq := octoprint.DeleteFileRequest{
			Location: octoprint.Local,
			Path:     filename,
		}

		// Perform the delete
		err := octoReq.Do(octoclient)
		if err != nil {
			return ("Error deleting file: " + err.Error())
		}

		return filename + " was deleted"

	} else {
		return "Not connected"
	}
}

func PrintOctoFile(commandList []string) string {
	switch len(commandList) {
	case 2:
		if commandList[1] == "-help" {
			return "usage: printfile filename.ext"
		}

		if !isConnected {
			return "Error: not connected to octoService"
		}

		// File in main/downloads
		fileName := commandList[1] // e.g., "Ring.gcode"

		// Destination in OctoPrint’s uploads
		user, err := user.Current()
		if err != nil {
			return fmt.Sprintf("Error getting user home: %v", err)
		}
		filePath := filepath.Join(user.HomeDir, ".octoprint/uploads", fileName)

		// Verify file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Sprintf("Error: file %s does not exist", filePath)
		}

		// Select and print file in OctoPrint’s local storage
		octoReq := octoprint.SelectFileRequest{
			Location: octoprint.Local,
			Path:     fileName, // Just the filename, as it’s in uploads/
			Print:    true,
		}

		err = octoReq.Do(octoclient)
		if err != nil {
			return "Error: not able to print file: " + err.Error()
		}
		return "File is printing"

	default:
		return "Error: parameter mismatch"
	}
}

/*
// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open source file
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Create destination file
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy contents
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// Ensure permissions
	return os.Chmod(dst, 0644) // rw-r--r--
}
*/

func SelectFile(c *octoprint.Client, filename string) string {
	if isConnected {
		octoReq := octoprint.SelectFileRequest{
			Location: octoprint.Local,
			Path:     filename,
			Print:    false,
		}

		err := octoReq.Do(c)
		if err != nil {
			return ("Error: selected file: " + err.Error())
		}
		return "File has been selected"
	} else {
		return "Error: not connected"
	}
}

func GetGcodeAnalysis(commandList []string) string {
	switch len(commandList) {
	case 2:
		if commandList[1] == "-help" {
			return "usage: printfile filename.ext"
		}

		if !isConnected {
			return "Error: not connected to octoService"
		}

		// Create file request
		octoReq := octoprint.FileRequest{
			Location:  octoprint.Local,
			Filename:  commandList[1], // e.g., "Ring.gcode"
			Recursive: false,
		}

		// Execute request
		response, err := octoReq.Do(octoclient)
		if err != nil {
			log.Printf("GetFileInfo error for %s: %v", commandList[1], err)
			return "Error: failed to get file info: " + err.Error()
		}

		analysis := GcodeAnalysis{
			EstimatedPrintTime: response.GCodeAnalysis.EstimatedPrintTime,
			Filament: struct {
				Length uint32  `json:"length"`
				Volume float64 `json:"volume"`
			}{
				Length: response.GCodeAnalysis.Filament.Length,
				Volume: response.GCodeAnalysis.Filament.Volume,
			},
		}

		jsonBytes, err := json.Marshal(analysis)
		if err != nil {
			return "Error: failed to encode analysis: " + err.Error()
		}
		return string(jsonBytes)

	default:
		return "Error: parameter mismatch"
	}
}

func GetJobStatus() string {
	if !isConnected {
		return "Error: not connected"
	}

	octoReq := octoprint.JobRequest{}
	job, err := octoReq.Do(octoclient)
	if err != nil {
		return ("Error GetJobStatus: " + err.Error())
	}

	status := JobStatus{
		FileName:    job.Job.File.Name,
		Progress:    math.Round(job.Progress.Completion),
		TimeElapsed: math.Round(job.Progress.PrintTime),
		TimeLeft:    math.Round(job.Progress.PrintTimeLeft),
	}

	jsonBytes, err := json.Marshal(status)
	if err != nil {
		return "Error: failed to encode job state: " + err.Error()
	}
	return string(jsonBytes)
}

func GetTemperature() string {
	if isConnected {
		return _getTemp()
	} else {
		return "Error: not connected"
	}
}

func _getTemp() string {
	octoReq := octoprint.StateRequest{}
	s, err := octoReq.Do(octoclient)
	if err != nil {
		return ("Error: " + err.Error())
	}

	var strB strings.Builder
	for tool, state := range s.Temperature.Current {
		fmt.Fprintf(&strB, "- %s: %.1f°C / %.1f°C\n", tool, state.Actual, state.Target)
	}

	str := strB.String()
	return str
}

func _getTempEntry() string {
	octoReq := octoprint.StateRequest{}
	s, err := octoReq.Do(octoclient)
	if err != nil {
		return ("Error: " + err.Error())
	}

	var results []TempState

	for tool, state := range s.Temperature.Current {
		result := TempState{
			Tool:   tool,
			Time:   time.Now().Format("2006-01-02 15:04:05.999999999 -0700 MST"),
			Actual: state.Actual,
			Target: state.Target,
		}
		results = append(results, result)
	}

	// Marshal the string directly into JSON format
	str, err := json.Marshal(results)
	if err != nil {
		return "Error: " + err.Error()
	}

	// fmt.Println("Data str: ", string(str))

	// Combine "addtemprecord" with the JSON string
	return "addtemprecord " + string(str)
}

func GetPrinterState() string {
	if !isConnected {
		return "Error: not connected to octoService"
	}

	// Get printer state
	printerReq := octoprint.StateRequest{}
	printer, err := printerReq.Do(octoclient)
	if err != nil {
		return "Error: failed to get printer state: " + err.Error()
	}

	// Get job state
	jobReq := octoprint.JobRequest{}
	job, err := jobReq.Do(octoclient)
	if err != nil {
		return "Error: failed to get job state: " + err.Error()
	}

	// Get current temperatures	-> remove this after we confirm this works!
	tempBytes, _ := json.Marshal(printer.Temperature)
	log.Printf("Raw current temperatures: %s", tempBytes)

	status := PrinterStatus{
		FileName: job.Job.File.Name,                   // e.g., "Ring.gcode"
		Progress: math.Round(job.Progress.Completion), // e.g., 75.2

		TimeLeft:     math.Round(job.Progress.PrintTimeLeft), // e.g., 1200
		State:        printer.State.Text,                     // e.g., "Printing"
		ExtruderTemp: 0.0,                                    // e.g., 210.0
		BedTemp:      0.0,                                    // e.g., 60.0

		Printing: printer.State.Flags.Printing,
	}

	// Iterate Current map for temps
	for tool, tempData := range printer.Temperature.Current {
		if tool == "tool0" {
			status.ExtruderTemp = tempData.Actual // e.g., 210.0
		}
		if tool == "bed" {
			status.BedTemp = tempData.Actual // e.g., 60.0
		}
	}

	// Encode to JSON
	jsonBytes, err := json.Marshal(status)
	if err != nil {
		return "Error: failed to encode printer state: " + err.Error()
	}

	return string(jsonBytes)
}

// Global tracker instance
var positionTracker *PositionTracker

// Initialize tracker at program start
func init() {
	var err error
	positionTracker, err = NewPositionTracker("/home/rich/.octoprint/logs/serial.log")
	if err != nil {
		fmt.Printf("Failed to initialize position tracker: %v\n", err)
	}
}

func GetPrintStatus() string {
	if !isConnected {
		return "Error: not connected"
	}

	// Fetch job data
	octoReq := octoprint.JobRequest{}
	job, err := octoReq.Do(octoclient)
	if err != nil {
		return "Error GetJobStatus: " + err.Error()
	}

	// Fetch printer state
	stateReq := octoprint.StateRequest{}
	state, err := stateReq.Do(octoclient)
	if err != nil {
		return "Error GetStateStatus: " + err.Error()
	}

	// Get position from tracker
	pos := positionTracker.GetPosition()

	// Build status
	status := PrintStatus{
		FileName:     job.Job.File.Name,
		Progress:     math.Round(job.Progress.Completion),
		TimeElapsed:  math.Round(job.Progress.PrintTime),
		TimeLeft:     math.Round(job.Progress.PrintTimeLeft),
		State:        state.State.Text,
		PositionX:    pos.X,
		PositionY:    pos.Y,
		PositionZ:    pos.Z,
		PositionE:    pos.E,
		ExtruderTemp: getTemperature(state, "tool0"),
		BedTemp:      getTemperature(state, "bed"),
	}

	jsonBytes, err := json.Marshal(status)
	if err != nil {
		return "Error: failed to encode print state: " + err.Error()
	}
	return string(jsonBytes)
}

/*
// getPosition sends M114 and parses response
func getPosition(client *octoprint.Client) (Position, error) {
	var pos Position

	// Try up to 2 times
	for attempt := 1; attempt <= 2; attempt++ {
		resp, err := client.CommandResponse([]string{"M114"})
		if err != nil {
			return pos, fmt.Errorf("failed to send M114 (attempt %d): %v", attempt, err)
		}

		// Parse logs for M114 output (e.g., "X:100.5 Y:75.2 Z:0.4")
		re := regexp.MustCompile(`X:([\d.]+)\s+Y:([\d.]+)\s+Z:([\d.]+)`)
		for _, log := range resp["logs"] {
			matches := re.FindStringSubmatch(log)
			if len(matches) == 4 {
				pos.X, _ = strconv.ParseFloat(matches[1], 64)
				pos.Y, _ = strconv.ParseFloat(matches[2], 64)
				pos.Z, _ = strconv.ParseFloat(matches[3], 64)
				return pos, nil
			}
		}

		// If no valid response, retry after a short delay
		if attempt < 2 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return pos, fmt.Errorf("no valid M114 response found in logs after 2 attempts")
}
*/

// getTemperature extracts temperature from FullStateResponse
func getTemperature(state *octoprint.FullStateResponse, tool string) float64 {
	if temp, exists := state.Temperature.Current[tool]; exists {
		return temp.Actual
	}
	return 0.0
}
