package agent

import (
	"net"
	"os"
)

// DEBUG is set to true, lots of print statement
// comes alive
var DEBUG bool = false

// AgentStat represent a single target
type Agent struct {
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

type File struct {
	CreateTime       int32   `json:"create_time"`
	FileName         string  `json:"fname"`
	FileSize         int64   `json:"fsize"`
	Id               int     `json:"id"`
	IsEnabled        bool    `json:"is_enabled"`
	IsPaused         bool    `json:"is_paused"`
	MimeType         string  `json:"mime_type"`
	Name             string  `json:"name"`
	OriginalMimeType string  `json:"orig_mime_type"`
	RedirectPath     string  `json:"redirect_path"`
	RefSubFile       int     `json:"ref_sub_file"`
	SubFile          *string `json:"sub_file"`
	SubMimeType      *string `json:"sub_mime_type"`
	SubName          *string `json:"sub_name"`
	Uid              int     `json:"uid"`
	UrlPath          string  `json:"url_path"`
}

type FileListData struct {
	Uploads []File `json:"uploads"`
}

type FileList struct {
	Data      FileListData `json:"data"`
	ErrorCode int          `json:"error_code"`
	Message   string       `json:"message"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GetLocalIP return their IP
// I say local because the agent might be behind a NAT network
// And their external IP is gonna be different.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "nil"
}
