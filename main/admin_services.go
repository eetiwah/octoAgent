package main

import (
	"encoding/json"

	"cwtch.im/cwtch/protocol/connections"
)

func Admin(commandList []string) string {
	switch len(commandList) {
	case 1:
		return "admin"

	case 2:
		if commandList[1] == "-help" {
			return "usage: admin"
		} else {
			return "Error: syntax, second attribute not recognized"
		}

	default:
		return "Error: parameter mismatch"
	}
}

func AddAdmin(commandList []string) string {
	switch len(commandList) {
	case 1:
		return "Error: missing adminNtKID"

	case 2:
		if commandList[1] == "-help" {
			return "usage: addadmin admin_ntkID"
		} else {
			newadmin := commandList[1]
			bot_admin_list = append(bot_admin_list, newadmin)
			return "Admin was added"
		}

	default:
		return "Error: parameter mismatch"
	}
}

func GetAdminList(commandList []string) string {
	switch len(commandList) {
	case 1:
		return getList(bot_admin_list)

	case 2:
		if commandList[1] == "-help" {
			return "usage: getadminlist"
		} else {
			return "Error: parameter mismatch"
		}

	default:
		return "Error: parameter mismatch"
	}
}

func RemoveAdmin(commandList []string) string {
	switch len(commandList) {
	case 1:
		return "Error: missing adminNtKID"

	case 2:
		if commandList[1] == "-help" {
			return "usage: removeadmin admin_ntkID"
		} else {
			admin := commandList[1]
			bot_admin_list = removeUserFromList(bot_admin_list, admin)
			return "Admin was removed"
		}

	default:
		return "Error: parameter mismatch"
	}
}

func AddContact(commandList []string) string {
	switch len(commandList) {
	case 1:
		return "Error: missing contactNtKID"

	case 2:
		if commandList[1] == "-help" {
			return "usage: addcontact contact_ntkID (admin, peer, user)"
		} else {
			return "Error: parameter mismatch"
		}

	case 3:
		newcontact := commandList[1]
		result := cwtchbot.Peer.ImportBundle(newcontact)
		if result != nil {
			return "Error: contact was not added"
		} else {
			switch commandList[2] {
			case "admin":
				bot_admin_list = append(bot_admin_list, newcontact)
				return "Contact was added and appended to bot_admin_list"

			case "peer":
				bot_peer_list = append(bot_peer_list, newcontact)
				return "Contact was added and appended to bot_peer_list"

			case "user":
				bot_user_list = append(bot_user_list, newcontact)
				return "Contact was added and appended to bot_user_list"

			default:
				return "Error: usertype not found"
			}
		}

	default:
		return "Error: parameter mismatch"
	}
}

func GetContactStatus(commandList []string) string {
	switch len(commandList) {
	case 1:
		return "Error: missing contactNtKID"

	case 2:
		if commandList[1] == "-help" {
			return "usage: contactstatus contact_ntkID (admin, peer, user)"
		} else {
			connectionState := cwtchbot.Peer.GetPeerState(commandList[1])
			switch connectionState {
			case connections.DISCONNECTED:
				return "Peer is not online"

			case connections.AUTHENTICATED:
				return "Peer is online"

			case connections.CONNECTED:
				return "Peer is online"

			case connections.CONNECTING:
				return "Peer is connecting"

			default:
				return "Peer state is unknown"
			}
		}

	default:
		return "Error: parameter mismatch"
	}
}

func AddPeer(commandList []string) string {
	switch len(commandList) {
	case 1:
		return "Error: missing peerNtKID"

	case 2:
		if commandList[1] == "-help" {
			return "usage: addapeer peer_ntkID"
		} else {
			newpeer := commandList[1]
			bot_peer_list = append(bot_peer_list, newpeer)
			return "Peer was added"
		}

	default:
		return "Error: parameter mismatch"
	}
}

func GetPeerList(commandList []string) string {
	switch len(commandList) {
	case 1:
		return getList(bot_peer_list)

	case 2:
		if commandList[1] == "-help" {
			return "usage: getpeerlist"
		} else {
			return "Error: parameter mismatch"
		}

	default:
		return "Error: parameter mismatch"
	}
}

func RemovePeer(commandList []string) string {
	switch len(commandList) {
	case 1:
		return "Error: missing peerNtKID"

	case 2:
		if commandList[1] == "-help" {
			return "usage: removepeer peer_ntkID"
		} else {
			peer := commandList[1]
			bot_peer_list = removeUserFromList(bot_peer_list, peer)
			return "Peer was removed"
		}

	default:
		return "Error: parameter mismatch"
	}
}

func AddUser(commandList []string) string {
	switch len(commandList) {
	case 1:
		return "Error: missing userNtKID"

	case 2:
		if commandList[1] == "-help" {
			return "usage: adduser user_ntkID"
		} else {
			newuser := commandList[1]
			bot_user_list = append(bot_user_list, newuser)
			return "User was added"
		}

	default:
		return "Error: parameter mismatch"
	}
}

func GetUserList(commandList []string) string {
	switch len(commandList) {
	case 1:
		return getList(bot_user_list)

	case 2:
		if commandList[1] == "-help" {
			return "usage: getuserlist"
		} else {
			return "Error: parameter mismatch"
		}

	default:
		return "Error: parameter mismatch"
	}
}

func RemoveUser(commandList []string) string {
	if len(commandList) == 2 {
		user := commandList[1]
		bot_user_list = removeUserFromList(bot_user_list, user)
		return "User was removed"
	} else {
		return "Error: parameter mismatch"
	}
}

func Help() string {
	return "Help"
}

func Ping() string {
	return "Pong"
}

func getList(list []string) string {
	if len(list) > 0 {
		jsonBytes, err := json.Marshal(list)
		if err != nil {
			return "Error: encoding of list"
		} else {
			return string(jsonBytes)
		}
	} else {
		return "The list was empty"
	}
}

func removeUserFromList(usersList []string, userToRemove string) []string {
	for i, user := range usersList {
		if user == userToRemove {
			// Remove the user from the slice by slicing it to exclude the user
			return append(usersList[:i], usersList[i+1:]...)
		}
	}
	// If userToRemove is not found, return the original slice
	return usersList
}
