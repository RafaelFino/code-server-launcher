package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Config struct {
	OAuthConfig   *oauth2.Config
	ProxmoxConfig *ProxmoxConfig
	GithubUrl     string
	UserListUrl   string
}

type ProxmoxConfig struct {
	Host          string
	Node          string
	Username      string
	Password      string
	TemplateID    int
	TimeToCloneVM int
}

func NewConfig(clientID, clientSecret, redirectURL string, githubUrl string, userListUrl string, proxmoxHost string, proxmoxNode string, proxmoxUsername string, proxmoxPassword string, proxmoxTemplateID int) *Config {
	oauthConf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	return &Config{
		OAuthConfig: oauthConf,
		GithubUrl:   githubUrl,
		UserListUrl: userListUrl,
		ProxmoxConfig: &ProxmoxConfig{
			Host:          proxmoxHost,
			Node:          proxmoxNode,
			Username:      proxmoxUsername,
			Password:      proxmoxPassword,
			TemplateID:    proxmoxTemplateID,
			TimeToCloneVM: 10,
		},
	}
}
