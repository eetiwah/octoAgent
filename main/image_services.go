package main

import (
	"os/exec"
	"time"
)

const (
	STILL      = "raspistill"
	VIDEO      = "raspivid"
	IMAGE_TYPE = ".jpg"
	VIDEO_TYPE = ".mp4"
	TIME_STAMP = "2006-01-02_15:04:05"
)

func TakePicture() string {
	// Define filename
	filename := bot_name + "-" + time.Now().Format(TIME_STAMP) + IMAGE_TYPE

	// Define command
	cmd := exec.Command(STILL, "-o", filename, "-w", "1280", "-h", "720", "-q", "80")

	// Run command
	err := cmd.Run()

	// Check for errors
	if err != nil {
		return "Error capturing image: " + err.Error()
	}

	return "Image captured and saved as: " + filename
}

func TakeVideo() string {
	// Define raspivid command
	raspividCmd := exec.Command(VIDEO, "-o", "-", "-t", "15000", "-w", "768", "-h", "480")

	// Define filename
	filename := bot_name + "-" + time.Now().Format(TIME_STAMP) + VIDEO_TYPE

	// Define ffmpeg command
	ffmpegCmd := exec.Command("ffmpeg", "-r", "30", "-i", "-", "-c:v", "copy", filename)

	// Pipe the output of raspivid to the input of ffmpeg
	ffmpegCmd.Stdin, _ = raspividCmd.StdoutPipe()

	// Start the raspivid command
	if err := raspividCmd.Start(); err != nil {
		return "Error starting raspivid: " + err.Error()
	}

	// Start the ffmpeg command
	if err := ffmpegCmd.Start(); err != nil {
		return "Error starting ffmpeg: " + err.Error()
	}

	// Wait for both commands to finish
	if err := raspividCmd.Wait(); err != nil {
		return "Error waiting for raspivid: " + err.Error()
	}
	if err := ffmpegCmd.Wait(); err != nil {
		return "Error waiting for ffmpeg: " + err.Error()
	}

	return "Video captured and saved as: " + filename
}
