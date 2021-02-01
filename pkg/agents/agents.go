package agents

import (
	"DiscordGo/pkg/agent"
)
var Agents []*agent.AgentInfo

func AddNewAgent(agentID string, hostname string, agentIP string, agentOS string){
	newAgent := &agent.Agent{}
	newAgent.UUID = agentID
	newAgent.HostName = hostname
	newAgent.IP = agentIP
	newAgent.OS = agentOS

	stat := &agent.AgentInfo{}
	stat.Agent = newAgent
	stat.Status = "alive"

	Agents = append(Agents, stat)
}

func RemoveAgent(agentID string){
	newAgentlist := []*agent.AgentInfo{}

	for _, agent := range Agents {
		if agent.Agent.UUID != agentID {
			newAgentlist = append(newAgentlist, agent)
		}
	}

	Agents = newAgentlist
}

func DoesAgentExist(agentID string) bool{
	// nihilism := false
	for _, agent := range Agents {
		if agent.Agent.UUID == agentID {
			return true
		}
	}
	return false
}