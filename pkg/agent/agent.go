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