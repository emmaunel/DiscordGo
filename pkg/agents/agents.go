package agents

import (
	"errors"

	"github.com/emmaunel/DiscordGo/pkg/agent"
)

var Agents []*agent.Agent

func AddNewAgent(agentID string, hostname string, agentIP string, agentOS string) {
	newAgent := &agent.Agent{}
	newAgent.UUID = agentID
	newAgent.HostName = hostname
	newAgent.IP = agentIP
	newAgent.OS = agentOS
	newAgent.Status = "Alive"
	newAgent.Timestamp = "null"

	// stat := &agent.AgentInfo{}
	// stat.Agent = newAgent
	// stat.Status = "alive"

	Agents = append(Agents, newAgent)
}

func RemoveAgent(agentID string) {
	newAgentlist := []*agent.Agent{}

	for _, agent := range Agents {
		if agent.UUID != agentID {
			newAgentlist = append(newAgentlist, agent)
		}
	}

	Agents = newAgentlist
}

func DoesAgentExist(agentID string) bool {
	// nihilism := false
	for _, agent := range Agents {
		if agent.UUID == agentID {
			return true
		}
	}
	return false
}

func UseAgent(agentID string) (agent.Agent, error) {

	for _, agent := range Agents {
		if agent.UUID == agentID {
			return *agent, nil
		}
	}
	emptyAgent := &agent.Agent{}
	return *emptyAgent, errors.New("couldn't not find agent")
}
