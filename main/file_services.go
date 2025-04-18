package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

const (
	DownloadFilePath = "./downloads"
	UpLoadFilePath   = "./uploads/"
)

/*
func Deletefile(commandList []string) string {
	switch len(commandList) {
	case 2:
		if commandList[1] == "-help" {
			return "usage: deletefile filename.ext"
		}

		return DeleteOctoFile(commandList[1])

	default:
		return "Error: parameter mismatch"
	}
}
*/

/*
func Getfilelist() string {
	// Destination in OctoPrint’s uploads
	user, err := user.Current()
	if err != nil {
		return fmt.Sprintf("Error getting user home: %v", err)
	}
	dirPath := filepath.Join(user.HomeDir, ".octoprint/uploads")

	// Verify directory exists
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			log.Printf("Directory %s does not exist", dirPath)
			return fmt.Sprintf("Error: directory %s does not exist", dirPath)
		}
		log.Printf("Error accessing directory %s: %v", dirPath, err)
		return fmt.Sprintf("Error: cannot access directory %s: %v", dirPath, err)
	}

	// Read directory
	fileInfos, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("Error reading directory %s: %v", dirPath, err)
		return "Error reading directory: " + err.Error()
	}

	var filelist []string

	// Create file list
	for _, file := range fileInfos {
		if !file.IsDir() {
			filelist = append(filelist, file.Name())
		}
	}

	// Encode file list
	if len(filelist) > 0 {
		jsonBytes, err := json.Marshal(filelist)
		if err != nil {
			log.Printf("Error encoding file list: %v", err)
			return "Error: encoding of list"
		} else {
			return string(jsonBytes)
		}
	} else {
		return "The directory was empty"
	}
}
*/

// This is an upload to S3 and then provide meta data for sharing
func Getfile(commandList []string) string {
	switch len(commandList) {
	case 2:
		if commandList[1] == "-help" {
			return packageReply("usage: getfile directory filename")
		} else {
			return packageReply("Error: syntax, second attribute not recognized")
		}

	case 3:
		// Set vars
		dir := commandList[1]
		filename := commandList[2]

		// Set decrypted file path
		decFilepath := SetFilePath(dir, filename)

		// Generate hash of decrypted file
		hash, err := FileHash(decFilepath)
		if err != nil {
			return packageReply("Error: could not hash the file")
		}

		// Encrypt file
		password := "Test"
		encFilepath := decFilepath + ".aes"
		err = EncryptFile(password, decFilepath, encFilepath)
		if err != nil {
			return packageReply("Error: could not encrypt file")
		}

		// Upload file
		result, err := s3UploadFile(AWS_S3_BUCKET, encFilepath)
		if err != nil {
			return result
		}

		// Create QBlock
		var block QBlock
		block.Location = "S3"
		block.Uri = AWS_HTTP_PREFIX + result

		// Create QBlockList
		var blockList []QBlock
		blockList = append(blockList, block)

		// create QShare
		var share QShare
		share.Actionabletype = "QShare"
		share.Sharetype = "File"
		share.Sharepassword = password
		share.Hash = hash
		share.Threshold = 0
		share.Shareblocks = blockList

		// Encode QShare struct
		shareJson, err := json.Marshal(share)
		if err != nil {
			return packageReply("Error: getfile json.Marshal of QShare")
		}

		return packageActionableReply(string(shareJson))

	default:
		return packageReply("Error: parameter mismatch")
	}
}

func Downloadfile(commandList []string) string {
	switch len(commandList) {
	case 1:
		return packageReply("Error: parameter mismatch")

	case 2:
		if commandList[1] == "-help" {
			return packageReply("usage: downloadfile s3FilePath")
		}

		// Assign URI
		downloadUri := commandList[1]

		// Split the downloadUri
		downloadUriList := strings.Split(downloadUri, "/")

		// Get the file name
		fileName := downloadUriList[len(downloadUriList)-1]

		// Destination in OctoPrint’s uploads
		user, err := user.Current()
		if err != nil {
			return fmt.Sprintf("Error getting user home: %v", err)
		}
		filePath := filepath.Join(user.HomeDir, ".octoprint/uploads", fileName)

		// Is the file already local, if not, go get it
		if FileExists(filePath) {
			fmt.Println("Cache hit: ", fileName)
			return filePath
		} else {
			fmt.Println("Retrieving file @ ", downloadUri)
			err := s3DownloadFile(filePath, downloadUri)
			if err != nil {
				return err.Error()
			}

			fmt.Println(fileName + " was downloaded")
			return fileName + " was downloaded"
		}

	default:
		return packageReply("Error: parameter mismatch")
	}
}

func Uploadfile(commandList []string) string {
	switch len(commandList) {
	case 2:
		if commandList[1] == "-help" {
			return "usage: getfile directory filename"
		} else {
			filepath := filepath.Join(UpLoadFilePath, commandList[1])
			if FileExists(filepath) {
				uri, err := s3UploadFile(AWS_S3_BUCKET, filepath)
				if err != nil {
					return "Error: not able to upload file"
				}

				var block QBlock
				block.Location = "S3"
				block.Uri = AWS_HTTP_PREFIX + uri
				blockJson, err := json.Marshal(block)
				if err != nil {
					return "Error: upload json.Marshal of QBlock"
				}

				// The return value must be contained in a list
				jsonStr := string(blockJson)
				var s3ShardList []string
				s3ShardList = append(s3ShardList, jsonStr)
				baList, _ := json.Marshal(s3ShardList)
				list := string(baList)
				return list
			} else {
				return "Error: file path does not exist"
			}
		}
	default:
		return "Error: parameter mismatch"
	}
}

func DirExists(dirPath string) bool {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.Mkdir(dirPath, 0755)
		if err != nil {
			fmt.Println("Error: cannot create directory: ", dirPath)
			return false
		}

		return true
	}
	return false
}

func FileExists(filepath string) bool {
	// Does the file exist?
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Filepath does not exist: ", filepath)
			return false
		} else {
			// fmt.Println(err)
			return false
		}
	} else {
		fmt.Println("Filepath exists: ", filepath)
		return true
	}
}

func SetDirectoryPath(dir string) string {
	if dir == "downloads" {
		return DownloadFilePath
	} else {
		return UpLoadFilePath
	}
}

func SetFilePath(dir string, filename string) string {
	if dir == "downloads" {
		return filepath.Join(DownloadFilePath, filename)
	} else {
		return filepath.Join(UpLoadFilePath, filename)
	}
}

func FileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write(fileContent)
	sha256Hash := hash.Sum(nil)

	// Convert the hash to a hexadecimal string representation
	hashStr := fmt.Sprintf("%x", sha256Hash)
	return hashStr, nil
}
