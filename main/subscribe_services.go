package main

import (
	"fmt"
	"strconv"
	"time"

	"cwtch.im/cwtch/model"
	"cwtch.im/cwtch/protocol/connections"
)

func Subscribe(commandList []string, id int) string {
	switch len(commandList) {
	case 2:
		if commandList[1] == "-help" {
			return "usage: subscribe increment duration [destinationNtKID]"
		} else {
			return "Error: syntax, second attribute not recognized"
		}

	case 3:
		// Convert increment from string to int
		increment, err := strconv.Atoi(commandList[1])
		if err != nil {
			return "Error: increment conversion failed "
		}

		// Convert duration from string to int
		duration, err := strconv.Atoi(commandList[2])
		if err != nil {
			return "Error: duration conversion failed "
		}

		// Set tick interval
		tickInterval := time.Duration(increment) * time.Second

		// Create a ticker that ticks every tickInterval
		ticker := time.NewTicker(tickInterval)

		// Define the function to be executed on each tick
		tickerFunc := func() {
			result := _getTempEntry()
			msg := string(cwtchbot.PackMessage(model.OverlayChat, result))
			cwtchbot.Peer.SendMessage(id, msg)
		}

		// Start a goroutine to execute the function on each tick
		go func() {
			tickerCount := 0
			for range ticker.C {
				// Execute the function on each tick
				tickerFunc()

				// Increment the ticketcounter
				tickerCount++

				// Are we done?
				if tickerCount*int(tickInterval/time.Second) >= duration {
					// Stop the ticker after duration seconds
					ticker.Stop()

					// Print log message
					fmt.Println("Ticker stopped after", duration, "seconds.")

					// Create completion message
					msg := string(cwtchbot.PackMessage(model.OverlayChat, "Subscribe completed"))

					// Send message to originator of subscription request
					cwtchbot.Peer.SendMessage(id, msg)
					break
				}
			}
		}()
		return "Subscribed!"

	case 4:
		// Convert increment from string to int
		increment, err := strconv.Atoi(commandList[1])
		if err != nil {
			return "Error: increment conversion failed "
		}

		// Convert duration from string to int
		duration, err := strconv.Atoi(commandList[2])
		if err != nil {
			return "Error: duration conversion failed "
		}

		// Get conversation information with destination
		conversation, _ := cwtchbot.Peer.FetchConversationInfo(commandList[3])

		// Ensure destination is online
		connectionState := cwtchbot.Peer.GetPeerState(commandList[3])

		if connectionState == connections.DISCONNECTED {
			return "Error: peer is not online"
		}

		// Set tick interval
		tickInterval := time.Duration(increment) * time.Second

		// Create a ticker that ticks every tickInterval
		ticker := time.NewTicker(tickInterval)

		// Define the function to be executed on each tick
		tickerFunc := func() {
			//			fmt.Println("Ticker ticked!") // Replace this statement with an action
			//			msg := string(cwtchbot.PackMessage(model.OverlayChat, "Ticker ticked!"))
			result := _getTempEntry()
			msg := string(cwtchbot.PackMessage(model.OverlayChat, result))
			cwtchbot.Peer.SendMessage(conversation.ID, msg)
		}

		// Start a goroutine to execute the function on each tick
		go func() {
			tickerCount := 0
			for range ticker.C {
				// Execute the function on each tick
				tickerFunc()

				// Increment the ticketcounter
				tickerCount++

				// Are we done?
				if tickerCount*int(tickInterval/time.Second) >= duration {
					// Stop the ticker after duration seconds
					ticker.Stop()

					// Print log message
					fmt.Println("Ticker stopped after", duration, "seconds.")

					// Create completion message
					msg := string(cwtchbot.PackMessage(model.OverlayChat, "Subscribe completed"))

					// Send message to originator of subscription request
					cwtchbot.Peer.SendMessage(id, msg)
					break
				}
			}
		}()
		return "Subscribed!"

	default:
		return "Error: parameter mismatch"

	}
}
