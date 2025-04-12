package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	TextMessageOverlay = 1

	ActionableMessageOverlay = 5
)

type OverlayEnvelope struct {
	onion   int
	Overlay int    `json:"o"`
	Data    string `json:"d"`
}

func adminMessages(envelope *OverlayEnvelope) string {
	switch envelope.Overlay {
	case TextMessageOverlay:
		cmd := strings.Split(envelope.Data, " ")
		id := envelope.onion
		fmt.Printf("\nCommand was received: %s\n", strings.ToLower(cmd[0]))

		switch strings.ToLower(cmd[0]) {
		case "admin":
			result := Admin(cmd)
			return packageReply(result)

		case "ping":
			result := Ping()
			return packageReply(result)

		case "help":
			result := Help()
			return packageReply(result)

		// Admin operations

		case "addadmin":
			result := AddAdmin(cmd)
			return packageReply(result)

		case "getadminlist":
			result := GetAdminList(cmd)
			return packageReply(result)

		case "removeadmin":
			result := RemoveAdmin(cmd)
			return packageReply(result)

		// Contact Operations

		case "addcontact":
			result := AddContact(cmd)
			return packageReply(result)

		case "contactstatus":
			result := GetContactStatus(cmd)
			return packageReply(result)

		// Peer Operations

		case "addpeer":
			result := AddPeer(cmd)
			return packageReply(result)

		case "getpeerlist":
			result := GetPeerList(cmd)
			return packageReply(result)

		case "removepeer":
			result := RemovePeer(cmd)
			return packageReply(result)

		// User Operations

		case "adduser":
			result := AddUser(cmd)
			return packageReply(result)

		case "getuserlist":
			result := GetUserList(cmd)
			return packageReply(result)

		case "removeuser":
			result := RemoveUser(cmd)
			return packageReply(result)

		// File operations

		case "deletefile":
			result := Deletefile(cmd)
			return packageReply(result)

		case "getfilelist":
			fileList := Getfilelist()
			return packageReply(fileList)

		case "getfile":
			return Getfile(cmd)

		case "downloadfile":
			result := Downloadfile(cmd)
			return packageReply(result)

		case "uploadfile":
			result := Uploadfile(cmd)
			return packageReply(result)

		// Image & Picture Operations

		case "takepicture":
			result := TakePicture()
			return packageReply(result)

		case "takevideo":
			result := TakeVideo()
			return packageReply(result)

		// Route and Subscribe Operations

		case "route":
			result := Route(cmd)
			return packageReply(result)

		case "subscribe":
			result := Subscribe(cmd, id)
			return packageReply(result)

		// *** Octo Service operations *** //

		case "getapiversion":
			result := GetApiVersion(cmd)
			return packageReply(result)

		case "checkoctoservice":
			result := checkOctoService()
			return packageReply(result)

		case "startoctoservice":
			result := StartOctoService()
			return packageReply(result)

		case "connectoctoservice":
			result := ConnectOctoService()
			return packageReply(result)

		case "disconnect":
			result := DisconnectOctoService()
			return packageReply(result)

		case "getconnection":
			result := GetConnection()
			return packageReply(result)

		case "printfile":
			result := PrintFile(cmd)
			return packageReply(result)

		case "getjobstatus":
			result := GetJobStatus(cmd)
			return packageReply(result)

		case "getprinterstate":
			result := GetPrinterState(cmd)
			return packageReply(result)

		case "gettemp":
			result := GetTemperature(cmd)
			return packageReply(result)

		default:
			return packageReply("Error: unrecognized command")
		}

	case ActionableMessageOverlay:
		// Get the command
		cmd := strings.Split(envelope.Data, " ")

		// Unmarshal JSON into QShare struct
		var qshare QShare
		err := json.Unmarshal([]byte(cmd[0]), &qshare)
		if err != nil {
			return "Error: " + err.Error()
		}

		switch qshare.Actionabletype {
		case "QShare":
			result := ProcessQShare(qshare)
			return packageReply(result)

		default:
			return packageReply("Error: unrecognized actionable message")
		}

	default:
		return packageReply("Error: unrecognized command")
	}
}

func peerMessages(envelope *OverlayEnvelope) {
	if envelope.Overlay == 1 {
		cmd := strings.Split(envelope.Data, " ")
		//		id := envelope.onion
		fmt.Printf("Peer command was received: %s\n", strings.ToLower(cmd[0]))

		switch cmd[0] {
		case "peer":
			fmt.Println("Peer message received")

		default:
			fmt.Printf("Error: unrecognized command = %v\n", cmd[0])
		}
	} else {
		fmt.Printf("unknown overlay type %d\n", envelope.Overlay)
	}
}

func userMessages(envelope *OverlayEnvelope) string {
	if envelope.Overlay == 1 {
		cmd := strings.Split(envelope.Data, " ")
		//		id := envelope.onion
		fmt.Printf("User command was received: %s\n", strings.ToLower(cmd[0]))

		switch cmd[0] {
		case "user":
			return packageReply("user")

		case "hello":
			return packageReply("Hello")

		case "help":
			return packageReply("Help")

		case "temp":
			return packageReply("Temp")

		default:
			return packageReply("Error: unrecognized command")
		}
	} else {
		fmt.Printf("unknown overlay type %d", envelope.Overlay)
		return packageReply("Error: unrecognized overlay")
	}
}
