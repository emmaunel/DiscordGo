package agent

// AgentStat represent a single target
type Agent struct {
	UUID      string
	HostName  string
	OS        string
	IP        string
	Status    string
	Timestamp string
}

// AgentInfo keeps track of each agent
type AgentInfo struct {
	Agent  *Agent
	Status string
}
