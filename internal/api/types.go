package api

import "fmt"

// Repo is the minimal Woodpecker repository shape.
type Repo struct {
	ID               int64  `json:"id"`
	ForgeRemoteID    string `json:"forge_remote_id,omitempty"`
	Owner            string `json:"owner"`
	Name             string `json:"name"`
	FullName         string `json:"full_name"`
	Avatar           string `json:"avatar_url,omitempty"`
	URL              string `json:"url,omitempty"`
	SCM              string `json:"scm,omitempty"`
	HTTPURL          string `json:"clone_url,omitempty"`
	SSHURL           string `json:"ssh_url,omitempty"`
	DefaultBranch    string `json:"default_branch,omitempty"`
	Timeout          int64  `json:"timeout,omitempty"`
	Visibility       string `json:"visibility,omitempty"`
	Private          bool   `json:"private,omitempty"`
	Trusted          bool   `json:"trusted,omitempty"`
	Protected        bool   `json:"protected,omitempty"`
	Active           bool   `json:"active,omitempty"`
	AllowPull        bool   `json:"allow_pull_requests,omitempty"`
	CancelPrev       bool   `json:"cancel_previous_pipeline_events,omitempty"`
	NetrcOnlyTrusted bool   `json:"netrc_only_trusted,omitempty"`
	OrgID            int64  `json:"org_id,omitempty"`
}

// Pipeline is the minimal Woodpecker pipeline shape.
type Pipeline struct {
	ID           int64    `json:"id"`
	RepoID       int64    `json:"repo_id"`
	Number       int64    `json:"number"`
	Parent       int64    `json:"parent,omitempty"`
	Event        string   `json:"event"`
	Status       string   `json:"status"`
	Enqueued     int64    `json:"enqueued_at,omitempty"`
	Created      int64    `json:"created_at,omitempty"`
	Started      int64    `json:"started_at,omitempty"`
	Finished     int64    `json:"finished_at,omitempty"`
	DeployTo     string   `json:"deploy_to,omitempty"`
	Commit       string   `json:"commit,omitempty"`
	Branch       string   `json:"branch,omitempty"`
	Ref          string   `json:"ref,omitempty"`
	RefSpec      string   `json:"refspec,omitempty"`
	Remote       string   `json:"remote,omitempty"`
	Title        string   `json:"title,omitempty"`
	Message      string   `json:"message,omitempty"`
	Timestamp    int64    `json:"timestamp,omitempty"`
	Sender       string   `json:"sender,omitempty"`
	Author       string   `json:"author,omitempty"`
	Email        string   `json:"email,omitempty"`
	LinkURL      string   `json:"link_url,omitempty"`
	ChangedFiles []string `json:"changed_files,omitempty"`
}

// PipelineConfig is the compiled pipeline YAML returned by the config endpoint.
type PipelineConfig struct {
	Data string `json:"data"`
}

// PipelineMetadata is the metadata associated with a pipeline.
type PipelineMetadata map[string]any

// Step is a workflow step in a pipeline.
type Step struct {
	ID         int64  `json:"id"`
	UUID       string `json:"uuid"`
	PID        int64  `json:"pid"`
	PPID       int64  `json:"ppid"`
	PipelineID int64  `json:"pipeline_id"`
	RepoID     int64  `json:"repo_id"`
	Name       string `json:"name"`
	State      string `json:"state"`
	ExitCode   int    `json:"exit_code,omitempty"`
	Started    int64  `json:"started,omitempty"`
	Stopped    int64  `json:"stopped,omitempty"`
}

// Log is a single log line.
type Log struct {
	ID     int64  `json:"id"`
	StepID int64  `json:"step_id"`
	Time   int64  `json:"time"`
	Line   int64  `json:"line"`
	Type   string `json:"type,omitempty"`
	Data   []byte `json:"data"`
}

// LogLineData is the common base64-encoded data in a Log.
type LogLineData struct {
	Out string `json:"out,omitempty"`
}

// User is the minimal Woodpecker user shape.
type User struct {
	ID     int64  `json:"id"`
	Login  string `json:"login"`
	Email  string `json:"email,omitempty"`
	Avatar string `json:"avatar_url,omitempty"`
	Admin  bool   `json:"admin,omitempty"`
	Active bool   `json:"active,omitempty"`
	Synced int64  `json:"synced,omitempty"`
}

// Agent is a Woodpecker agent.
type Agent struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Token       string `json:"token,omitempty"`
	LastContact int64  `json:"last_contact,omitempty"`
	Platform    string `json:"platform,omitempty"`
	Backend     string `json:"backend,omitempty"`
	Capacity    int64  `json:"capacity,omitempty"`
	NoSchedule  bool   `json:"no_schedule,omitempty"`
	Version     string `json:"version,omitempty"`
}

// QueueInfo is the queue status response.
type QueueInfo struct {
	Stats struct {
		Workers       int  `json:"workers"`
		Pending       int  `json:"pending"`
		WaitingOnDeps int  `json:"waiting_on_deps"`
		Running       int  `json:"running"`
		Total         int  `json:"total"`
		Paused        bool `json:"paused,omitempty"`
	} `json:"stats"`
}

// Cron is a scheduled pipeline.
type Cron struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	RepoID    int64  `json:"repo_id"`
	CreatorID int64  `json:"creator_id,omitempty"`
	NextExec  int64  `json:"next_exec,omitempty"`
	Schedule  string `json:"schedule,omitempty"`
	Branch    string `json:"branch,omitempty"`
}

// Secret is a Woodpecker secret.
type Secret struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Value  string   `json:"value,omitempty"`
	Images []string `json:"images,omitempty"`
	Events []string `json:"events,omitempty"`
}

// Registry is a container registry credential.
type Registry struct {
	ID       int64  `json:"id"`
	Address  string `json:"address"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Org is a Woodpecker organization.
type Org struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	IsUser  bool   `json:"is_user,omitempty"`
	Private bool   `json:"private,omitempty"`
}

// Version is the server version response.
type Version struct {
	Source  string `json:"source,omitempty"`
	Version string `json:"version,omitempty"`
}

// Forge is a forge (repository host) entry.
type Forge struct {
	ID           int64  `json:"id"`
	URL          string `json:"url"`
	Type         string `json:"type,omitempty"`
	Client       string `json:"client,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// APIError is a normalized upstream error.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// BadRequest returns true for 4xx client errors.
func (e APIError) BadRequest() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// Unauthorized returns true for 401.
func (e APIError) Unauthorized() bool {
	return e.StatusCode == 401
}

// Forbidden returns true for 403.
func (e APIError) Forbidden() bool {
	return e.StatusCode == 403
}

// NotFound returns true for 404.
func (e APIError) NotFound() bool {
	return e.StatusCode == 404
}

// ServerError returns true for 5xx errors.
func (e APIError) ServerError() bool {
	return e.StatusCode >= 500
}

// RepoNotFoundError is returned when an owner/repo cannot be resolved.
type RepoNotFoundError struct {
	FullName string
}

func (e RepoNotFoundError) Error() string {
	return fmt.Sprintf("repository %q not found", e.FullName)
}
