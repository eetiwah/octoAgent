package main

import (
	"errors"
	"fmt"
	"log"
	bot "octoAgent"
	"octoAgent/octoprint"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"

	"cwtch.im/cwtch/event"
	"cwtch.im/cwtch/model/attr"
	"cwtch.im/cwtch/model/constants"

	"github.com/joho/godotenv"
	_ "github.com/mutecomm/go-sqlcipher/v4"
)

// Define global variables
var (
	bot_name       = ""
	bot_attribute  = ""
	bot_admin_ID   = ""
	bot_admin_list []string
	bot_peer_list  []string
	bot_user_list  []string
	cwtchbot       *bot.CwtchBot

	// octoBot specific
	baseURL    = ""
	apiKey     = ""
	octoclient *octoprint.Client
)

func setGlobalVars() error {
	// Find .env file
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}

	// Set global vars
	bot_name = os.Getenv("NAME")
	bot_attribute = os.Getenv("ATTRIBUTE")
	bot_admin_ID = os.Getenv("ADMIN")
	baseURL = os.Getenv("BASE_URL")
	apiKey = os.Getenv("API_KEY")

	// Is the bot_name empty? If so, we got a problem!
	if bot_name == "" {
		log.Println("Error: bot_name is empty")
		return errors.New("bot_name is empty")
	}

	// Is the bot_admin_ID empty? If so, we got a problem!
	if bot_admin_ID == "" {
		log.Println("Error: bot_admin_ID is empty")
		return errors.New("bot_admin_ID is empty")
	}

	// Set bot_admin_list
	bot_admin_list = append(bot_admin_list, bot_admin_ID)
	return nil
}

func instantiateAgent() error {
	// Instantiate new agent
	botpath := "/" + bot_name + "/"

	switch runtime.GOOS {
	case "windows":
		cwtchbot = bot.NewCwtchBot(path.Join("./tor/win", botpath), bot_name)

	case "linux":
		_path := path.Join("./tor/linux", botpath)
		cwtchbot = bot.NewCwtchBot(_path, bot_name)

	default:
		return fmt.Errorf("operating system not support = %v", runtime.GOOS)
	}

	cwtchbot.Launch() // Need some error check here

	// Set Some Profile Information
	cwtchbot.Peer.SetScopedZonedAttribute(attr.PublicScope, attr.ProfileZone, constants.Name, bot_name)
	cwtchbot.Peer.SetScopedZonedAttribute(attr.PublicScope, attr.ProfileZone, constants.ProfileAttribute1, bot_attribute)
	cwtchbot.Peer.SetScopedZonedAttribute(attr.PublicScope, attr.ProfileZone, constants.ProfileAttribute2, bot_attribute)
	cwtchbot.Peer.SetScopedZonedAttribute(attr.PublicScope, attr.ProfileZone, constants.ProfileAttribute3, bot_attribute)

	// Display address
	log.Printf("%s address: %v\n", bot_name, cwtchbot.Peer.GetOnion())

	return nil
}

func main() {
	// Set global variables
	err := setGlobalVars()
	if err != nil {
		fmt.Printf("Error loading .env file: %s", err)
		return
	}

	// Create the bot
	err = instantiateAgent()
	if err != nil {
		log.Printf("Error: instantiating the bot, %s", err)
		return
	}

	// Initialize new octoclient
	octoclient = octoprint.NewClient(baseURL, apiKey)
	if octoclient == nil {
		log.Println("Error: octoclient not initialized")
		return
	}

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Processing loop
	go func() {
		log.Println("Starting message queue go coroutine")
		for {
			message := cwtchbot.Queue.Next()
			conversation, _ := cwtchbot.Peer.FetchConversationInfo(message.Data[event.RemotePeer])
			switch message.EventType {
			case event.ContactCreated:
				log.Printf("Received contact request from %v %v\n", conversation, message.Data[event.RemotePeer])

				if inList(conversation.Handle, bot_admin_list) {
					cwtchbot.Peer.AcceptConversation(conversation.ID)
					reply := packageReply("You been added as an admin")
					cwtchbot.Peer.SendMessage(conversation.ID, reply)
				} else if inList(conversation.Handle, bot_user_list) {
					cwtchbot.Peer.AcceptConversation(conversation.ID)
					reply := packageReply("You been added as an user")
					cwtchbot.Peer.SendMessage(conversation.ID, reply)
				} else {
					log.Printf("Contact request refused from %v %v\n", conversation, message.Data[event.RemotePeer])
				}

			case event.NewMessageFromPeer:
				log.Println("NewMessageFromPeer")
				log.Printf("Remote Peer = %v\n", message.Data[event.RemotePeer])
				log.Printf("Raw envelope = %v\n", message.Data[event.Data])

				envelope := Unwrap(conversation.ID, message.Data[event.Data])
				log.Printf("Data = %v\n", envelope.Data)

				// Check if this is a response or not
				if envelope.Data != "Error:" && envelope.Data != "Success" {
					if inList(conversation.Handle, bot_admin_list) {
						reply := adminMessages(envelope)
						cwtchbot.Peer.SendMessage(conversation.ID, reply)
					} else if inList(conversation.Handle, bot_user_list) {
						reply := userMessages(envelope)
						cwtchbot.Peer.SendMessage(conversation.ID, reply)
					} else {
						log.Printf("Error: contact does not have sufficient privileges, message from %v %v\n", conversation, message.Data[event.RemotePeer])
					}
				} else {
					// The response will be logged
					log.Printf("Response: %s\n", envelope.Data)
				}

			case event.PeerStateChange:
				data := message.Data
				log.Printf("PeerStateChange: %s: %s", data[event.ConnectionState], data[event.RemotePeer])

			case event.PeerAcknowledgement:
				log.Println("PeerAcknowledgement")
				log.Printf("Msg EventID = %v\n", message.EventID)
				log.Printf("Data EventID = %v\n", message.Data[event.EventID])
				log.Printf("Remote Peer = %v\n", message.Data[event.RemotePeer])

			case event.SendRetValMessageToPeer:
				// We need to dig into this, but it does not effect the functionality of the bot

			case event.NewGetValMessageFromPeer:
				// We need to dig into this, but it does not effect the functionality of the bot

			default:
				log.Printf("Unhandled event: %v\n", message.EventType)
			}
		}

	}()

	// Block until a signal is received
	<-shutdown
	log.Println("Shutting down gracefully...")
}
