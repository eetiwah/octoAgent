package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"cwtch.im/cwtch/model"
	"cwtch.im/cwtch/protocol/connections"
)

/*

Publish:

This function publishes temp information derived from _getTempEntry
every "increment" seconds until math.Round(status.Progress) is equal to 100 to peerID, where
peerId can represent a peer or a group formed by a communityAgent. The adminID
represents the admin or conductorAgent that is organizing/requesting the process.

Note: if peerID = adminID then _getTempEntry data will be sent to the adminID

getTempEntry provides the following information:
Tool:   tool,
Time:   time.Now().Format("2006-01-02 15:04:05.999999999 -0700 MST"),
Actual: state.Actual,
Target: state.Target,

*/

func Publish(commandList []string, adminID int) string {
	switch len(commandList) {
	case 2:
		if commandList[1] == "-help" {
			return "usage: publish increment peerId"
		} else {
			return "Error: syntax, second attribute not recognized"
		}

	case 3:
		// Convert increment from string to int
		increment, err := strconv.Atoi(commandList[1])
		if err != nil {
			return "Error: increment conversion failed "
		}

		// Get conversation information with destination
		conversation, err := cwtchbot.Peer.FetchConversationInfo(commandList[2])
		if err != nil {
			return "Error: peer is not a contact: " + commandList[2]
		}

		// Check peer online
		connectionState := cwtchbot.Peer.GetPeerState(commandList[2])
		if connectionState == connections.DISCONNECTED {
			return "Error: peer is not online"
		}

		// Set tick interval
		tickInterval := time.Duration(increment) * time.Second

		// Create a ticker that ticks every tickInterval
		ticker := time.NewTicker(tickInterval)

		// Define the function to be executed on each tick
		tickerFunc := func() {
			result := _getTempEntry()
			msg := string(cwtchbot.PackMessage(model.OverlayChat, result))
			if _, err := cwtchbot.Peer.SendMessage(conversation.ID, msg); err != nil {
				log.Printf("Error sending completion message to admin %d: %v", conversation.ID, err)
			} else {
				log.Printf("Sent completion message to admin %d", conversation.ID)
			}
		}

		// Start a goroutine to execute the function on each tick
		go func() {
			for range ticker.C {
				// Execute the function on each tick
				tickerFunc()

				// Get the Job Status to evaluate using it instead of duration for termination
				statusRsp := GetJobStatus()
				var status JobStatus
				err := json.Unmarshal([]byte(statusRsp), &status)
				if err != nil {
					fmt.Printf("Error: failed to encode job state: %v", err.Error())
					continue
				}

				// Log progress
				log.Printf("Job Progress: %.1f, TimeLeft: %.1f", status.Progress, status.TimeLeft)

				// Are we done?
				if math.Round(status.Progress) >= 100 {
					// Stop the ticker after duration seconds
					ticker.Stop()

					// Print log message
					fmt.Println("Print job is complete...")

					// Create completion message
					msg := string(cwtchbot.PackMessage(model.OverlayChat, "Publishing completed"))

					// Send message to originator of subscription request
					if _, err := cwtchbot.Peer.SendMessage(adminID, msg); err != nil {
						log.Printf("Error sending completion message to admin %d: %v", adminID, err)
					} else {
						log.Printf("Sent completion message to admin %d", adminID)
					}
					break
				}
			}
		}()
		return "Publishing has started"

	default:
		return "Error: parameter mismatch"

	}
}
