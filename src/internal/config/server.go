package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Config struct {
	OAuthConfig   *oauth2.Config `json:"oauth_config"`
	ProxmoxConfig *ProxmoxConfig `json:"proxmox_config"`
	GithubUrl     string         `json:"github_url"`
	UserListUrl   string         `json:"user_list_url"`
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

func NewProxmoxConfig(host, node, username, password string, vmTemplateID, memSize, cpuCores int, storageName string, storageSize int, baseIP string, networkInterface string, timeToStart int) *ProxmoxConfig {
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

func NewOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

func NewConfig(oauthConfig *oauth2.Config, proxmoxConfig *ProxmoxConfig, githubUrl string, userListUrl string) *Config {
	return &Config{
		OAuthConfig:   oauthConfig,
		ProxmoxConfig: proxmoxConfig,
		GithubUrl:     githubUrl,
		UserListUrl:   userListUrl,
	}
}
