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
	"strconv"
	"strings"
	"syscall"

	"DiscordGo/pkg/util/constants"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var channelID *discordgo.Channel
var fileInputPtr string
var list = make(map[string]Target)

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
	// var list = make(map[string]Target)
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

func removeDuplicatesValues(arrayToEdit []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range arrayToEdit {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func createOrDeleteRoles(dg *discordgo.Session, create bool) {
	log.Info("Creating Roles....")
	g, err := dg.Guild(constants.ServerID)
	if err != nil {
		log.Error("Something broke ", err)
		return
	}

	potentialRole := []string{} // Roles to be created
	availbleRoles := g.Roles    // Roles already created and listed from discord

	for _, host := range list {
		potentialRole = append(potentialRole, host.hostname)
		potentialRole = append(potentialRole, host.teamstring)
	}

	// New list without duplicates
	rolesToCreate := removeDuplicatesValues(potentialRole) // Roles to be created

	if !create { // We want to delete the roles
		for _, role := range availbleRoles {
			for _, roleToDelete := range potentialRole {
				log.Info("Deleting " + role.Name)
				if roleToDelete == role.Name || role.Name == "new role" {
					dg.GuildRoleDelete(constants.ServerID, role.ID)
					break
				}
			}
		}
	} else {
		// TODO Do a check if a role already exist
		for _, role := range rolesToCreate {
			// Color Fix: Thank Fred
			var colorInRGB randomcolor.RGBColor = randomcolor.GetRandomColorInRgb()
			roleColorHex := fmt.Sprintf("%.2x%.2x%.2x", colorInRGB.Red, colorInRGB.Green, colorInRGB.Blue)
			roleColorInt64, err := strconv.ParseInt(roleColorHex, 16, 64)
			if err != nil {
				log.Error(err)
			}
			roleColorInt := int(roleColorInt64)

			log.Info("Creating " + role + " role with color RGB: " + strconv.Itoa(roleColorInt))
			newRole, err := dg.GuildRoleCreate(constants.ServerID)
			if err != nil {
				log.Error(err)
			}
			// Editing the role template
			_, err = dg.GuildRoleEdit(constants.ServerID, newRole.ID, role, roleColorInt, false, 171429441, true)
			if err != nil {
				log.Error(err)
			}
		}
	}
	log.Info("Creating Roles Ended........")
}

func cleanChannels(dg *discordgo.Session, targetFile string) {
	log.Info("Start Clean")
	targetMap, teams, _ := parseCSV(targetFile)
	checkChannels, _ := dg.GuildChannels(constants.ServerID)
	for _, catName := range teams {
		groupExixsts := false
		for _, channelCheck := range checkChannels {
			if channelCheck.Name == catName {
				groupExixsts = true
				log.Warn("Category already exist")
				break
			}
		}
		if !groupExixsts {
			log.Info("Creating non-Existing group")
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
	log.Info("End Clean")
}

func main() {
	flag.StringVar(&fileInputPtr, "target", "", "This csv should contains the list of targets: ip,team#,hostname")
	statPtr := flag.Bool("stat", false, "True to create role, False to delete roles")
	flag.Parse()

	if fileInputPtr == "" {
		log.Fatal("No file specified")
		os.Exit(0)
	}

	log.SetOutput(os.Stdout)
	log.Info("Target file: " + fileInputPtr)

	dg, err := discordgo.New("Bot " + constants.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	log.Info("Bot Connected")

	cleanChannels(dg, fileInputPtr)

	// true --> Create roles
	// false --> Delete roles
	if *statPtr {
		createOrDeleteRoles(dg, false) // Change the value to false to delete roels
	} else {
		createOrDeleteRoles(dg, true) // Change the value to false to delete roels
	}

	dg.AddHandler(guimessageCreater)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
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

		log.Info("Looking at the channels")
		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildText {
				for _, hostname := range hostnames {
					if strings.ToLower(hostname) == channel.Name {
						log.Info("Deleting channel: " + channel.Name)
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
		log.Info("Looking at the category")
		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildCategory {
				for _, teamnum := range teamnums {
					if teamnum == channel.Name {
						log.Info("Deleting category: " + channel.Name)
						_, err := dg.ChannelDelete(channel.ID)
						if err != nil {
							log.Error(err)
						}
						break
					}

				}
			}
		}

	}
}
