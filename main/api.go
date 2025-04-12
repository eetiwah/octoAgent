package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"octoAgent/octoprint"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
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

type Progress struct {
	Completion    float64 `json:"completion"`
	Filepos       int     `json:"filepos"`
	PrintTime     int     `json:"printTime"`
	PrintTimeLeft int     `json:"printTimeLeft"`
}

type Payload struct {
	Job      Job      `json:"job"`
	Progress Progress `json:"progress"`
}

type File struct {
	Name          string        `json:"name"`
	Path          string        `json:"path"`
	Type          string        `json:"type"`
	TypePath      []string      `json:"typePath"`
	Hash          string        `json:"hash"`
	Size          int           `json:"size"`
	Date          int64         `json:"date"`
	Origin        string        `json:"origin"`
	Refs          Refs          `json:"refs"`
	GcodeAnalysis GcodeAnalysis `json:"gcodeAnalysis"`
	Print         Print         `json:"print"`
}

type Refs struct {
	Resource string `json:"resource"`
	Download string `json:"download"`
	Model    string `json:"model"`
}

type GcodeAnalysis struct {
	EstimatedPrintTime int `json:"estimatedPrintTime"`
	Filament           struct {
		Length int `json:"length"`
		Volume int `json:"volume"`
	} `json:"filament"`
}

type Print struct {
	Failure int64 `json:"failure"`
	Success int64 `json:"success"`
	Last    struct {
		Date    int64 `json:"date"`
		Success bool  `json:"success"`
	} `json:"last"`
}

type FilesResponse struct {
	Files []File `json:"Files"`
	Free  int    `json:"Free"`
}

func GetApiVersion(commandList []string) string {

	switch len(commandList) {
	case 1:
		octoReq := octoprint.VersionRequest{}
		s, err := octoReq.Do(octoclient)
		if err != nil {
			return ("Error: " + err.Error())
		}

		return s.API + "\n"

	case 2:
		if commandList[1] == "-help" {
			return "usage: getapiversion"
		} else {
			return "Error: syntax, second attribute not recognized"
		}

	default:
		return "Error: parameter mismatch"

	}
}

// Helper to check if OctoPrint is running
func checkOctoService() string {
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
	// Check if OctoPrint is running
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
		str := "Disconnected\n"
		return str
	} else {
		return "Already disconnected\n"
	}
}

func GetConnection() string {
	// Check if OctoPrint is running
	if !_checkOctoService() {
		return "Octoprint server is NOT already running, please start it"
	}

	if isConnected {
		octoReq := octoprint.ConnectionRequest{}
		s, err := octoReq.Do(octoclient)
		if err != nil {
			return ("Error: " + err.Error())
		}

		str := string(s.Current.State)
		return str + "\n"
	} else {
		return "Not connected"
	}
}

func GetFileInfo(commandList []string) string {
	switch len(commandList) {
	case 2:
		if commandList[1] == "-help" {
			return "usage: getfileinfo filename.ext"
		} else {
			if isConnected {
				octoReq := octoprint.FileRequest{
					Location:  octoprint.SDCard, // octoprint.Local
					Filename:  commandList[1],
					Recursive: false,
				}

				jsonResponse, err := octoReq.Do(octoclient)
				if err != nil {
					return ("Error GetFileInfo: " + err.Error())
				}

				// Convert jobResponse to JSON bytes
				jsonData, err := json.Marshal(jsonResponse)
				if err != nil {
					return "Error converting to JSON: " + err.Error()
				}

				// Create an instance of FilesResponse struct to hold the JSON data
				var file File

				// Unmarshal the JSON data into the FilesResponse struct
				if err := json.Unmarshal(jsonData, &file); err != nil {
					return "Error unmarshalling JSON: " + err.Error()
				}

				result := "Name: " + file.Name + "\n"
				result += "Size: " + strconv.Itoa(file.Size) + "\n"
				result += "Hash: " + file.Hash + "\n"
				result += "Location: " + file.Origin + "\n"
				result += "EstTime: " + strconv.Itoa(file.GcodeAnalysis.EstimatedPrintTime) + "\n"
				return result
			} else {
				return "Not connected\n"
			}
		}

	default:
		return "Error: parameter mismatch"
	}
}

func GetOctoFileList() string {
	if isConnected {
		octoReq := octoprint.FilesRequest{
			Location:  octoprint.SDCard, // octoprint.Local
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
			fmt.Fprintf(&strB, "Name: %s\nSize: %d\nEstTime: %d\n", file.Name, file.Size, file.GcodeAnalysis.EstimatedPrintTime)
		}

		str := strB.String()
		return str

	} else {
		return "Not connected\n"
	}
}

func AddFile(c *octoprint.Client, filename string, fileContent []byte) string {
	if isConnected {
		octoReq := octoprint.UploadFileRequest{
			Location: octoprint.SDCard,
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
		return "Not connected\n"
	}
}

func DeleteOctoFile(filename string) string {
	if isConnected {
		octoReq := octoprint.DeleteFileRequest{
			Location: octoprint.SDCard,
			Path:     filename,
		}

		// Perform the delete
		err := octoReq.Do(octoclient)
		if err != nil {
			return ("Error deleting file: " + err.Error())
		}

		return filename + " was deleted"

	} else {
		return "Not connected\n"
	}
}

func PrintFile(commandList []string) string {
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
		srcPath := filepath.Join("/home/rich/Projects/octo/octoAgent/main/downloads", fileName)

		// Verify source file exists
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			return fmt.Sprintf("Error: file %s does not exist", srcPath)
		}

		// Destination in OctoPrint’s uploads
		user, err := user.Current()
		if err != nil {
			return fmt.Sprintf("Error getting user home: %v", err)
		}
		destPath := filepath.Join(user.HomeDir, ".octoprint/uploads", fileName)

		// Copy file to OctoPrint’s uploads
		err = copyFile(srcPath, destPath)
		if err != nil {
			return fmt.Sprintf("Error copying file to %s: %v", destPath, err)
		}

		// Select and print file in OctoPrint’s local storage
		octoReq := octoprint.SelectFileRequest{
			Location: octoprint.Local,
			Path:     fileName, // Just the filename, as it’s in uploads/
			Print:    true,
		}

		err = octoReq.Do(octoclient)
		if err != nil {
			return "Error PrintFile: " + err.Error()
		}
		return "File is printing"

	default:
		return "Error: parameter mismatch"
	}
}

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

func SelectFile(c *octoprint.Client, filename string) string {
	if isConnected {
		octoReq := octoprint.SelectFileRequest{
			Location: octoprint.SDCard,
			Path:     filename,
			Print:    false,
		}

		err := octoReq.Do(c)
		if err != nil {
			return ("Error SelectFile: " + err.Error())
		}
		return "File has been selected\n"
	} else {
		return "Not connected\n"
	}
}

func GetJobStatus(commandList []string) string {
	switch len(commandList) {
	case 1:
		if isConnected {
			octoReq := octoprint.JobRequest{}
			jobResponse, err := octoReq.Do(octoclient)
			if err != nil {
				return ("Error GetJobStatus: " + err.Error())
			}

			// Convert jobResponse to JSON bytes
			jsonData, err := json.Marshal(jobResponse)
			if err != nil {
				return "Error converting to JSON: " + err.Error()
			}

			var payload Payload

			// Unmarshal the JSON data into the Payload struct
			if err := json.Unmarshal(jsonData, &payload); err != nil {
				return "Error unmarshalling JSON: " + err.Error()
			}

			// Format Completion field to one decimal point
			completionStr := fmt.Sprintf("%.1f", payload.Progress.Completion)

			// Build result
			result := "Job: " + payload.Job.File.Name + "\n"
			result += "Estimated completion: " + strconv.Itoa(payload.Job.EstimatedPrintTime) + "\n"
			result += "Completed(%): " + completionStr + "\n"
			result += "Elapsed time(s): " + strconv.Itoa(payload.Progress.PrintTime) + "\n"
			return result

		} else {
			return "Not connected\n"
		}

	case 2:
		if commandList[1] == "-help" {
			return "usage: getjobstatus"
		} else {
			return "Error: syntax, second attribute not recognized"
		}

	default:
		return "Error: parameter mismatch"
	}
}

func GetTemperature(commandList []string) string {
	switch len(commandList) {
	case 1:
		if isConnected {
			return _getTemp()
		} else {
			return "Not connected\n"
		}

	case 2:
		if commandList[1] == "-help" {
			return "usage: gettemp"
		} else {
			return "Error: syntax, second attribute not recognized"
		}

	default:
		return "Error: parameter mismatch"
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

type TempState struct {
	Tool   string
	Time   string
	Actual float64
	Target float64
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

	fmt.Println("Data str: ", string(str))

	// Combine "addtemprecord" with the JSON string
	return "addtemprecord " + string(str)
}

func GetPrinterState(commandList []string) string {
	return "Error: not implemented"

	/*	switch len(commandList) {
		case 1:
			if isConnected {
				octoReq := octoprint.PrinterState{}
				list, err := octoReq.Do(octoclient)
				if err != nil {
					return ("Error: " + err.Error())
				}

				jsonBytes, err := json.Marshal(list)
				if err != nil {
					return "Error: encoding of list"
				} else {
					return string(jsonBytes)
				}
			} else {
				return "Not connected"
			}

		case 2:
			if commandList[1] == "-help" {
				return "usage: gettemp"
			} else {
				return "Error: syntax, second attribute not recognized"
			}

		default:
			return "Error: parameter mismatch"
		}
	*/
}
