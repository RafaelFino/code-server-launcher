package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type AppConfig struct {
	Github      *GithubConfig  `json:"github"`
	Proxmox     *ProxmoxConfig `json:"proxmox"`
	Server      *ServerConfig  `json:"server"`
	UserListUrl string         `json:"user_list_url"`
}

type GithubConfig struct {
	oauth2.Config
	GithubUrl string `json:"github_url"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type ProxmoxConfig struct {
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

func NewGithubAuth(clientID, clientSecret, redirectURL, githubUrl string) *GithubConfig {
	return &GithubConfig{
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

func (g *GithubConfig) GetOAuth() *oauth2.Config {
	return &g.Config
}

func NewServer(host string, port int) *ServerConfig {
	return &ServerConfig{
		Host: host,
		Port: port,
	}
}

func NewProxmox(host, node, username, password string, vmTemplateID, memSize, cpuCores int, storageName string, storageSize int, baseIP string, networkInterface string, timeToStart int) *ProxmoxConfig {
	return &ProxmoxConfig{
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

func NewConfig(githubConfig *GithubConfig, proxmoxConfig *ProxmoxConfig, serverConfig *ServerConfig, githubUrl string, userListUrl string) *AppConfig {
	return &AppConfig{
		Github:      githubConfig,
		Proxmox:     proxmoxConfig,
		Server:      serverConfig,
		UserListUrl: userListUrl,
	}
}
