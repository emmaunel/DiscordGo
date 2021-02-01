package message

import (
	"DiscordGo/pkg/util/constants"
	"encoding/json"

	"github.com/bwmarrin/discordgo"
)

type Type int

const (
	MESSAGE_PING Type = iota
	MESSAGE_PONG Type = iota
	MESSAGE_OUTPUT Type = iota
	MESSAGE_COMMAND Type = iota
	MESSAGE_NEW Type = iota
	MESSAGE_KILL Type = iota
	MESSAGE_DISCONNECT Type = iota
)

type Message struct {
	AgentID     string
	MessageType Type
	Message 	string
	Server		bool
}

// NewMessage wiill create a new Message struct and convert it json
func NewMessage(agentID string, message string, fromServer bool, messageType Type) string{
	newMessage := &Message{}
	newMessage.AgentID = agentID
	newMessage.Message = message
	newMessage.MessageType = messageType
	newMessage.Server = fromServer

	// change to json
	messageJSON, _ := json.Marshal(newMessage)
	return string(messageJSON)
}

func CommandMessage(dg *discordgo.Session, agentID string, command string) {
	message := NewMessage(agentID, command, true, MESSAGE_COMMAND)
	dg.ChannelMessageSend(constants.ChannelID, message)

}

func Ping(dg *discordgo.Session, agentID string){
	message := NewMessage(agentID, "", true, MESSAGE_PING)
	dg.ChannelMessageSend(constants.ChannelID, message)
}

func KillAgent(dg *discordgo.Session, agentID string) {
	message := NewMessage(agentID, "", true, MESSAGE_KILL)
	dg.ChannelMessageSend(constants.ChannelID, message)
}
