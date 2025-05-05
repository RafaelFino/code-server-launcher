package service

import (
	"bytes"
	"code-server-launcher/internal/config"
	"code-server-launcher/internal/domain"
	"code-server-launcher/internal/logger"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

type Caddy struct {
	config.CaddyConfig
	log *logger.Logger
}

type ReverseProxyHandler struct {
	Handler   string `json:"handler"`
	Upstreams []struct {
		Dial string `json:"dial"`
	} `json:"upstreams"`
}

// Estrutura de rota HTTP
type Route struct {
	Match []struct {
		Host []string `json:"host"`
	} `json:"match"`
	Handle   []ReverseProxyHandler `json:"handle"`
	Terminal bool                  `json:"terminal"`
}

func NewCaddyService(cfg *config.CaddyConfig) *Caddy {
	return &Caddy{
		log:         logger.NewLogger("CaddyService"),
		CaddyConfig: *cfg,
	}
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func (c *Caddy) Insert(user *domain.User, internalIp string, internalPort int) error {
	subdomain := fmt.Sprintf("%s.%s", user.Login, c.BaseURL)
	internalIP := fmt.Sprintf("%s.%d", c.BaseInternalIP, user.ID)

	if !isValidIP(internalIP) {
		return fmt.Errorf("invalid internal IP: %s", internalIP)
	}

	route := Route{
		Match: []struct {
			Host []string `json:"host"`
		}{
			{Host: []string{subdomain}},
		},
		Handle: []ReverseProxyHandler{
			{
				Handler: "reverse_proxy",
				Upstreams: []struct {
					Dial string `json:"dial"`
				}{
					{Dial: fmt.Sprintf("%s:%d", internalIP, internalPort)},
				},
			},
		},
		Terminal: true,
	}

	jsonData, err := json.Marshal(route)
	if err != nil {
		c.log.Error("Failed to marshal JSON: %v", err)
		return err
	}

	caddyUrl := fmt.Sprintf("http://%s:%d/config/apps/http/servers/srv0/routes", c.Host, c.Port)
	resp, err := http.Post(caddyUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.log.Error("Failed to send request to Caddy: %v", err)
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.log.Error("Failed to add route to Caddy: %s", resp.Status)
		return fmt.Errorf("failed to add route to Caddy: %s", resp.Status)
	}

	c.log.Debug("Caddy response status: %s", resp.Status)
	return nil
}
