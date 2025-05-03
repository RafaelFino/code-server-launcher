package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Config struct {
	Github      *Github  `json:"github"`
	Proxmox     *Proxmox `json:"proxmox"`
	Server      *Server  `json:"server"`
	UserListUrl string   `json:"user_list_url"`
}

type Github struct {
	oauth2.Config
	GithubUrl string `json:"github_url"`
}

type Server struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Proxmox struct {
	Host             string `json:"host"`
	Node             string `json:"node"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	TemplateID       int    `json:"template_id"`
	MemSize          int    `json:"memory_size"`
	CPUCores         int    `json:"cpu_cores"`
	StorageName      string `json:"storage_name"`
	StorageSize      int    `json:"storage_size"`
	NetworkInterface string `json:"network_interface"`
	BaseIP           string `json:"base_ip"`
	TimetoStart      int    `json:"time_to_start"`
}

func NewGithubAuth(clientID, clientSecret, redirectURL, githubUrl string) *GithubAuthConfig {
	return &Github{
		GithubUrl: githubUrl,
		Config: oauth2.Config{
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
		},
	}
}

func (g *Github) GetOAuth() *oauth2.Config {
	return &g.Config
}

func NewServer(host string, port int) *Server {
	return &Server{
		Host: host,
		Port: port,
	}
}

func NewProxmox(host, node, username, password string, vmTemplateID, memSize, cpuCores int, storageName string, storageSize int, baseIP string, networkInterface string, timeToStart int) *Proxmox {
	return &Proxmox{
		Host:             host,
		Node:             node,
		Username:         username,
		Password:         password,
		TemplateID:       vmTemplateID,
		MemSize:          memSize,
		CPUCores:         cpuCores,
		StorageName:      storageName,
		StorageSize:      storageSize,
		BaseIP:           baseIP,
		NetworkInterface: networkInterface,
		TimetoStart:      timeToStart,
	}
}

func NewConfig(githubConfig *Github, proxmoxConfig *Proxmox, serverConfig *Server, githubUrl string, userListUrl string) *Config {
	return &Config{
		Github:      githubConfig,
		Proxmox:     proxmoxConfig,
		Server:      serverConfig,
		UserListUrl: userListUrl,
	}
}
