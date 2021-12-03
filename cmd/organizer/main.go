package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/emmaunel/DiscordGo/pkg/util/constants"

	"github.com/bwmarrin/discordgo"
)


var channelID *discordgo.Channel
var fileInputPtr string

type Target struct {
	ip         string
	teamstring string
	hostname   string
}

// PwnBoard json post request
type PwnBoard struct {
	IPs  string `json:"ip"`
	Type string `json:"type"`
}

func parseCSV(csvName string) (map[string]Target, []string, []string) {
	var list = make(map[string]Target)
	f, err := os.Open(csvName)
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(f)
	teamnum := []string{}
	hostnameList := []string{}
	for s.Scan() {
		lineBuff := s.Bytes()
		v := strings.Split(string(lineBuff), ",")
		list[v[0]] = Target{
			ip:         v[0],
			teamstring: v[1],
			hostname:   v[2],
		}
		teamnum = append(teamnum, v[1])
		hostnameList = append(hostnameList, v[2])
	}
	return list, teamnum, hostnameList
}

func cleanChannels(dg *discordgo.Session, targetFile string) {
	println(targetFile) // TODO Add to logging
	println("Start Clean")
	targetMap, teams, _ := parseCSV(targetFile)
	checkChannels, _ := dg.GuildChannels(constants.ServerID)
	for _, catName := range teams {
		groupExixsts := false
		for _, channelCheck := range checkChannels {
			if channelCheck.Name == catName {
				groupExixsts = true
				break
			}
		}
		if !groupExixsts {
			println("Creating non-Existing group")
			newChan, _ := dg.GuildChannelCreate(constants.ServerID, catName, 4)
			checkChannels = append(checkChannels, newChan)
		}
	}

	var channelName2ID = make(map[string]string)
	channels, _ := dg.GuildChannels(constants.ServerID)
	for _, channel := range channels {
		if _, ok := channelName2ID[channel.Name]; !ok {
			channelName2ID[channel.Name] = channel.ID
		}
	}
	for _, channel := range channels {
		if _, ok := channelName2ID[channel.Name]; !ok {
			channelName2ID[channel.Name] = channel.ID
		}
		if target, ok := targetMap[channel.Name]; ok {
			if group_id, ok := channelName2ID[target.teamstring]; ok {
				dg.ChannelEditComplex(channel.ID, &discordgo.ChannelEdit{ParentID: group_id, Name: target.hostname})
			}
		}
	}
	println("End Clean")

}

func main() {
	flag.StringVar(&fileInputPtr, "target", "", "This csv should contains the list of targets: ip,team#,hostname")
	flag.Parse()

	if fileInputPtr == "" {
		fmt.Println("No file specified")
		os.Exit(0)
	}

	dg, err := discordgo.New("Bot " + constants.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	cleanChannels(dg, fileInputPtr)

	dg.AddHandler(guimessageCreater)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		// fmt.Println("error opening connection,", err)
		return
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

// updatepwnBoard sends a post request to pwnboard with the IP
// Request is done every 15 seconds
// ip: Victim's IP
func updatepwnBoard(ip string) {
	url := "http://pwnboard.win/generic"

	data := PwnBoard{
		IPs:  ip,
		Type: "DiscordG0",
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

func guimessageCreater(dg *discordgo.Session, message *discordgo.MessageCreate) {
	if strings.HasPrefix(message.Content, "!heartbeat") {
		agent_ip_address := strings.Split(message.Content, " ")[1]
		updatepwnBoard(agent_ip_address)
	}

	if message.Author.ID == dg.State.User.ID {
		return
	}

	if message.Content == "export" {
		names := []string{}
		channels, _ := dg.GuildChannels(message.GuildID)
		for _, channel := range channels {
			names = append(names, channel.Name)
		}
		dg.ChannelMessageSend(message.ChannelID, "```"+strings.Join(names, "\n")+"```")
	}

	if message.Content == "clean" {
		cleanChannels(dg, fileInputPtr)
		dg.ChannelMessageSend(message.ChannelID, "Cleaned")
	}

	if message.Content == "delcomp" { 
		channels, _ := dg.GuildChannels(constants.ServerID)
		_, teamnums, hostnames := parseCSV(fileInputPtr)

		println("Looking at the channels")
		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildText {
				for _, hostname := range hostnames {
					if strings.ToLower(hostname) == channel.Name {
						println("Deleting channel: " + channel.Name)
						_, err := dg.ChannelDelete(channel.ID)
						if err != nil {
							println(err)
						}
						break
					}
				}
			}
		}

		println()
		println("Looking at the category")
		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildCategory {
				for _, teamnum := range teamnums {
					if teamnum == channel.Name {
						println("Deleting category: " + channel.Name)
						_, err := dg.ChannelDelete(channel.ID)
						if err != nil {
							println(err)
						}
						break
					}

				}
			}
		}

	}


}
