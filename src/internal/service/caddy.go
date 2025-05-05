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

func (c *Caddy) GetRoutes() ([]Route, error) {
	caddyUrl := fmt.Sprintf("http://%s:%d/config/apps/http/servers/srv0/routes", c.Host, c.Port)
	resp, err := http.Get(caddyUrl)
	if err != nil {
		c.log.Error("Failed to send request to Caddy: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Error("Failed to get routes from Caddy: %s", resp.Status)
		return nil, fmt.Errorf("failed to get routes from Caddy: %s", resp.Status)
	}

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		c.log.Error("Failed to decode JSON response: %v", err)
		return nil, err
	}

	return routes, nil
}

func (c *Caddy) ExistsRoute(user *domain.User) (bool, error) {
	routes, err := c.GetRoutes()
	if err != nil {
		return false, err
	}

	subdomain := fmt.Sprintf("%s.%s", user.Login, c.BaseURL)
	for _, route := range routes {
		if len(route.Match) > 0 && len(route.Match[0].Host) > 0 && route.Match[0].Host[0] == subdomain {
			c.log.Debug("Route already exists for user %s", user.Login)
			return true, nil
		}
	}

	c.log.Debug("Route does not exist for user %s", user.Login)
	return false, nil
}

func (c *Caddy) Insert(user *domain.User, internalIp string, internalPort int) error {
	subdomain := fmt.Sprintf("%s.%s", user.Login, c.BaseURL)
	internalIP := fmt.Sprintf("%s.%d", c.BaseInternalIP, user.ID)

	if !isValidIP(internalIP) {
		c.log.Error("Invalid internal IP: %s", internalIP)
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

	var routeExists bool
	routeExists, err = c.ExistsRoute(user)
	if err != nil {
		c.log.Error("Failed to check if route exists: %v", err)
		return err
	}

	var resp *http.Response
	method := "POST"

	if routeExists {
		method = "PUT"
	}

	req, err := http.NewRequest(method, caddyUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		c.log.Error("Failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)

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
