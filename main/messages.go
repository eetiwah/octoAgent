package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

const (
	TextMessageOverlay       = 1
	ActionableMessageOverlay = 5
	SuggestContactOverlay    = 100
	InviteGroupOverlay       = 101
)

type OverlayEnvelope struct {
	onion   int
	Overlay int    `json:"o"`
	Data    string `json:"d"`
}

type GroupInvite struct {
	GroupID   string `json:"GroupID"`
	GroupName string `json:"GroupName"`
	//	SignedGroupID string `json:"SignedGroupID"` // Nullable string
	Timestamp  int    `json:"Timestamp"`
	SharedKey  string `json:"SharedKey"`
	ServerHost string `json:"ServerHost"`
}

// Message holds parts data
type Message struct {
	O int    `json:"o"`
	D string `json:"d"`
}

func adminMessages(envelope *OverlayEnvelope) string {
	cmd := strings.Split(envelope.Data, " ")
	requestingPeerId := envelope.onion
	fmt.Printf("Command was received: %s", strings.ToLower(cmd[0]))

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

	// Upload/Download operations
	/*
		case "deletefile":
			result := Deletefile(cmd)
			return packageReply(result)
	*/
	/*
		case "getfilelist":
			fileList := Getfilelist()
			return packageReply(fileList)
	*/
	/*
		case "getfile":
			return Getfile(cmd)
	*/
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

	case "publish":
		result := Publish(cmd, requestingPeerId)
		return packageReply(result)

	// *** Octo Service operations *** //

	case "getapiversion":
		result := GetApiVersion()
		return packageReply(result)

	case "checkoctoservice":
		result := CheckOctoService()
		return packageReply(result)

	case "startoctoservice":
		result := StartOctoService()
		return packageReply(result)

	case "connectoctoservice":
		result := ConnectOctoService()
		return packageReply(result)

	case "disconnectoctoservice":
		result := DisconnectOctoService()
		return packageReply(result)

	case "getconnectionsettings":
		result := GetConnectionSettings()
		return packageReply(result)

	case "deleteoctofile":
		result := DeleteOctoFile(cmd)
		return packageReply(result)

	case "getoctofilelist":
		result := GetOctoFileList()
		return packageReply(result)

	case "getoctofileinfo":
		result := GetOctoFileInfo(cmd)
		return packageReply(result)

	case "printoctofile":
		result := PrintOctoFile(cmd)
		return packageReply(result)

	case "getgcodeanalysis":
		result := GetGcodeAnalysis(cmd)
		return packageReply(result)

	case "getjobstatus":
		result := GetJobStatus()
		return packageReply(result)

	case "getprinterstate":
		result := GetPrinterState()
		return packageReply(result)

	case "gettemp":
		result := GetTemperature()
		return packageReply(result)

	default:
		return packageReply("Error: unrecognized command")
	}
}

func inviteGroup(bundle string) string {
	// Is this a valid bundle
	err := decodeGroupInvite(bundle)
	if err != nil {
		return packageReply(err.Error())
	}

	err = cwtchbot.Peer.ImportBundle(bundle)
	if err != nil {
		return packageReply("Error: Import Bundle: " + err.Error())
	}

	return packageReply("Group Invite was successful")
}

// decodeGroupInvite processes a cwtch group invite
func decodeGroupInvite(content string) error {
	// Log raw input
	log.Printf("\n\nRaw content: %q (length: %d)", content, len(content))

	// Split on ||
	parts := strings.Split(strings.TrimSpace(content), "||")
	log.Printf("Parts: %v (length: %d)", parts, len(parts))
	if len(parts) != 2 {
		return fmt.Errorf("error: invalid content: expected 2 parts, got %d", len(parts))
	}

	// Skip first 5 characters of parts[1]
	if len(parts[1]) <= 5 {
		return fmt.Errorf("error: parts[1] too short: %d chars", len(parts[1]))
	}
	encoded := parts[1][5:]
	log.Printf("Encoded Base64: %q (length: %d)", encoded, len(encoded))

	// Validate Base64
	if !regexp.MustCompile(`^[A-Za-z0-9+/=]+$`).MatchString(encoded) {
		return fmt.Errorf("error: invalid Base64 characters in encoded string")
	}

	// Decode Base64
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// Try RawStdEncoding for padding issues
		log.Printf("Error: StdEncoding failed: %v, trying RawStdEncoding", err)
		decodedBytes, err = base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return fmt.Errorf("error: base64 decode failed: %v", err)
		}
	}
	log.Printf("Decoded string: %s", string(decodedBytes))

	// Parse JSON
	var invite GroupInvite
	if err := json.Unmarshal(decodedBytes, &invite); err != nil {
		return fmt.Errorf("error: json decode failed: %v", err)
	}

	// This is just for testing purposes
	fmt.Printf("GroupID = %s\n", invite.GroupID)
	fmt.Printf("GroupName = %s\n", invite.GroupName)
	fmt.Printf("ServerHost = %s\n", invite.ServerHost)
	fmt.Printf("SharedKey = %s\n", invite.SharedKey)

	return nil
}

/*
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
*/

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
