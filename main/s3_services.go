package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	_ "github.com/joho/godotenv/autoload"
)

var (
	// This is only needed if we are uploading to S3
	AWS_S3_REGION   = os.Getenv("AWS_S3_REGION")
	AWS_AccessKey   = os.Getenv("AWS_AccessKey")
	AWS_Secret      = os.Getenv("AWS_Secret")
	AWS_HTTP_PREFIX = os.Getenv("AWS_HTTP_PREFIX")
	AWS_S3_BUCKET   = os.Getenv("AWS_S3_BUCKET")
	token           = ""
)

func s3DownloadFile(filepath string, url string) error {
	// Does directory exists?	--> this is repeated in file_services.go => clean up!!!
	dir := "downloads"
	if err := ensureDir(dir); err != nil {
		return err
	}

	// Create the filepath	--> this is repeated in file_services.go => clean up!!!
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	// Get content on file
	resp, err := client.Get(url)
	if err != nil {
		// log.Error(err)
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		// fmt.Printf("bad status: %s", resp.Status)
		return errors.New(resp.Status)
	}

	// Writer the body to file
	size, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Printf("Downloaded file to: %s with size %d\n", filepath, size)
	return nil
}

// ensureDir checks if a directory exists, creates it if not
func ensureDir(dirPath string) error {
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Create directory with 0755 perms (rwxr-xr-x)
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
		}
		log.Printf("Created directory: %s", dirPath)
	} else if err != nil {
		return fmt.Errorf("error checking directory %s: %v", dirPath, err)
	} else {
		log.Printf("Directory already exists: %s", dirPath)
	}
	return nil
}

func s3UploadFile(bucketPath string, fpath string) (string, error) {
	// Generate S3 credentials
	creds := credentials.NewStaticCredentials(AWS_AccessKey, AWS_Secret, token)
	_, err := creds.Get()
	if err != nil {
		fmt.Println("Error: could not generate AWS credentials")
		return "", err
	}

	// Set S3 configuration
	cfg := aws.NewConfig().WithRegion(AWS_S3_REGION).WithCredentials(creds)

	// Create session using configuration
	session, err := session.NewSession(cfg)
	if err != nil {
		fmt.Println("Error: could not create AWS session")
		return "", err
	}

	// Open file
	uploadFile, err := os.Open(fpath)
	if err != nil {
		fmt.Println("Error: could not open upload file")
		return "", err
	}
	defer uploadFile.Close()

	// Get filename
	filename := filepath.Base(fpath)

	// Get upload file info
	upFileInfo, _ := uploadFile.Stat()
	var fileSize int64 = upFileInfo.Size()

	// Read file into buffer
	fileBuffer := make([]byte, fileSize)
	uploadFile.Read(fileBuffer)

	// Set S3 params
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucketPath),
		Key:           aws.String(filename),
		ACL:           aws.String("public-read"),
		Body:          bytes.NewReader(fileBuffer),
		ContentLength: aws.Int64(fileSize),
		ContentType:   aws.String(http.DetectContentType(fileBuffer)),
	}

	// Upload file
	_, err = s3.New(session).PutObject(params)
	if err != nil {
		return "", err
	} else {
		uploadedFilePath := bucketPath + filename
		return uploadedFilePath, nil
	}
}
