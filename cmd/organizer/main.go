package main

import (
	"bufio"
	"bytes"
	"encoding/json"
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

func parseCSV(csvName string) (map[string]Target, []string) {
	var list = make(map[string]Target)
	f, err := os.Open(csvName)
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(f)
	blH := []string{}
	for s.Scan() {
		lineBuff := s.Bytes()
		v := strings.Split(string(lineBuff), ",")
		list[v[0]] = Target{
			ip:         v[0],
			teamstring: v[1],
			hostname:   v[2],
		}
		blH = append(blH, v[1])
	}
	return list, blH
}

func cleanChannels(dg *discordgo.Session) {

	println("Start Clean")
	targetMap, teams := parseCSV("targets_UB.csv")
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
	dg, err := discordgo.New("Bot " + constants.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	cleanChannels(dg)

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
		cleanChannels(dg)
		dg.ChannelMessageSend(message.ChannelID, "Cleaned")
	}
}
