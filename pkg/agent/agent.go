package agent

// AgentStat represent a single target
type Agent struct {
	UUID	 string
	HostName string
	OS 		 string
	IP       string // This will be the local IP
	EIP		 string // This will be the external IP
}

// AgentInfo keeps track of each agent
type AgentInfo struct {
	Agent *Agent
	Status string
}