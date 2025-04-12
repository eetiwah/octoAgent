package main

import (
	"cwtch.im/cwtch/model"
	"cwtch.im/cwtch/protocol/connections"
)

// Command structure: route msg ntkID
func Route(commandList []string) string {
	switch len(commandList) {
	case 2:
		if commandList[1] == "-help" {
			return "usage: route destinationNtKID messagetype [message]"
		} else {
			return "Error: syntax, second attribute not recognized"
		}

	case 3:
		// Get conversation information with destination
		conversation, _ := cwtchbot.Peer.FetchConversationInfo(commandList[1])

		// Get connection state of the NtKID
		connectionState := cwtchbot.Peer.GetPeerState(commandList[1])

		// Ensure destination is online
		if connectionState == connections.DISCONNECTED {
			return "Error: peer is not online"
		}

		// Create message to be routed
		message := commandList[2]
		msg := string(cwtchbot.PackMessage(model.OverlayChat, message))

		// Send message to destination
		_, err := cwtchbot.Peer.SendMessage(conversation.ID, msg)
		if err != nil {
			return "Error: message was not routed"
		} else {
			return "Message was routed"
		}

	case 4:
		// Get conversation information with destination
		conversation, _ := cwtchbot.Peer.FetchConversationInfo(commandList[1])

		// Get connection state of the NtKID
		connectionState := cwtchbot.Peer.GetPeerState(commandList[1])

		// Ensure destination is online
		if connectionState == connections.DISCONNECTED {
			return "Error: peer is not online"
		}

		// Create message to be routed
		message := commandList[2] + " " + commandList[3]
		msg := string(cwtchbot.PackMessage(model.OverlayChat, message))

		// Send message to destination
		_, err := cwtchbot.Peer.SendMessage(conversation.ID, msg)
		if err != nil {
			return "Error: message was not routed"
		} else {
			return "Message was routed"
		}

	default:
		return "Error: parameter mismatch"

	}
}
