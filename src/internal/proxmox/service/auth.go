/*package service

import (
	"bytes"
	"code-server-launcher/internal/logger"
	"code-server-launcher/internal/proxmox/config"
	"code-server-launcher/internal/proxmox/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Auth struct {
	Username string
	Password string
	host     string
	log      *logger.Logger
	Session  *domain.ProxmoxSession
}

func NewAuthService(config *config.ProxmoxConfig) *Auth {
	return &Auth{
		host:     config.Host,
		Username: config.Username,
		Password: config.Password,
		log:      logger.NewLogger("ProxmoxAuthService"),
		Session:  &domain.ProxmoxSession{},
	}
}

func (a *Auth) CheckClusterStatus() error {
	url := fmt.Sprintf("%s/cluster/status", a.host)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		p.log.Error("Failed to create request: %v", err)
		return err
	}

	req.Header.Set("CSRFPreventionToken", a.Session.CSRFPreventionToken)
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", a.Session.Ticket))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		p.log.Error("Failed to make request: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		p.log.Warn("Failed to get cluster status: %s, maybe need to renew session ticket", resp.Status)
		return fmt.Errorf("failed to get cluster status: %s, maybe need to renew session ticket", resp.Status)
	}

	return nil
}

func (a *Auth) GetSession(header *http.Header) error {
	err := a.CheckClusterStatus()

	defer func(session *domain.ProxmoxSession, header *http.Header) {
		header.Set("CSRFPreventionToken", session.CSRFPreventionToken)
		header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", session.Ticket))
		header.Set("Content-Type", "application/json")
	}(a.Session, header)

	if a.Session.Ticket != "" && a.Session.CSRFPreventionToken != "" && err == nil {
		p.log.Debug("Ticket and CSRF token already set, skipping authentication")
		return nil
	}

	url := a.host + "/access/ticket"

	loginData := map[string]string{
		"Username": a.Username,
		"Password": a.Password,
		"realm":    "pam",
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		p.log.Error("Failed to marshal login data: %v", err)
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		p.log.Error("Failed to make POST request: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.log.Error("[Failed to read response body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		p.log.Error("Failed to authenticate: %s", body)
		return fmt.Errorf("failed to authenticate: %s", body)
	}

	var authResp domain.AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		p.log.Error("Failed to unmarshal response: %v", err)
		return err
	}

	p.log.Debug("Authentication successful, ticket: %s, CSRF token: %s", authResp.Data.Ticket, authResp.Data.CSRFPreventionToken)

	p.Session.Ticket = authResp.Data.Ticket
	p.Session.CSRFPreventionToken = authResp.Data.CSRFPreventionToken

	return nil
}
*/