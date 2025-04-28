package service

import (
	"bytes"
	"code-server-launcher/internal/code-server-laucher/config"
	"code-server-launcher/internal/code-server-laucher/domain"
	"code-server-launcher/internal/code-server-laucher/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ProxmoxService struct {
	Host                string
	Node                string
	Username            string
	Password            string
	TemplateID          int
	Ticket              string
	CSRFPreventionToken string
	hostUrl             string
	TimeToCloneVM       time.Duration
}

var log = logger.Logger

type AuthResponse struct {
	Data struct {
		Ticket              string `json:"ticket"`
		CSRFPreventionToken string `json:"CSRFPreventionToken"`
	} `json:"data"`
}

func NewProxmoxService(config *config.Config) *ProxmoxService {
	ret := &ProxmoxService{
		Host:          config.ProxmoxConfig.Host,
		Node:          config.ProxmoxConfig.Node,
		Username:      config.ProxmoxConfig.Username,
		Password:      config.ProxmoxConfig.Password,
		TemplateID:    config.ProxmoxConfig.TemplateID,
		TimeToCloneVM: time.Duration(config.ProxmoxConfig.TimeToCloneVM) * time.Second,
	}

	ret.hostUrl = "https://" + ret.Host + "/api2/json"

	return ret
}

func (p *ProxmoxService) checkClusterStatus() error {
	url := fmt.Sprintf("%s/cluster/status", p.hostUrl)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("[ProxmoxService.checkClusterStatus] Failed to create request: %v", err)
		return err
	}

	req.Header.Set("CSRFPreventionToken", p.CSRFPreventionToken)
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", p.Ticket))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("[ProxmoxService.checkClusterStatus] Failed to make request: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("[ProxmoxService.checkClusterStatus] Failed to get cluster status: %s", resp.Status)
		return fmt.Errorf("failed to get cluster status: %s", resp.Status)
	}

	return nil
}

func (p *ProxmoxService) getTicket() error {
	err := p.checkClusterStatus()

	if p.Ticket != "" && p.CSRFPreventionToken != "" && err == nil {
		log.Debug("[ProxmoxService.getTicket] Ticket and CSRF token already set, skipping authentication")
		return nil
	}

	url := p.hostUrl + "/access/ticket"

	loginData := map[string]string{
		"username": p.Username,
		"password": p.Password,
		"realm":    "pam",
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		log.Errorf("[ProxmoxService.getTicket] Failed to marshal login data: %v", err)
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("[ProxmoxService.getTicket] Failed to make POST request: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[ProxmoxService.getTicket] Failed to read response body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("[ProxmoxService.getTicket] Failed to authenticate: %s", body)
		return fmt.Errorf("failed to authenticate: %s", body)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		log.Errorf("[ProxmoxService.getTicket] Failed to unmarshal response: %v", err)
		return err
	}

	log.Debug("[ProxmoxService.getTicket] Authentication successful, ticket: %s, CSRF token: %s", authResp.Data.Ticket, authResp.Data.CSRFPreventionToken)
	p.Ticket = authResp.Data.Ticket
	p.CSRFPreventionToken = authResp.Data.CSRFPreventionToken

	return nil
}

type VMStatusResponse struct {
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
}

func (p *ProxmoxService) getVMStatus(user *domain.User) (string, error) {
	url := fmt.Sprintf("%s/nodes/%s/qemu/%s/status/current", p.hostUrl, p.Node, user.ID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("[ProxmoxService.getVMStatus] Failed to create request: %v", err)
		return "", err
	}

	err = p.getTicket()
	if err != nil {
		log.Errorf("[ProxmoxService.getVMStatus] Failed to get ticket: %v", err)
		return "", err
	}
	req.Header.Set("CSRFPreventionToken", p.CSRFPreventionToken)
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", p.Ticket))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("[ProxmoxService.getVMStatus] Failed to make request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[ProxmoxService.getVMStatus] Failed to read response body: %v", err)
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("[ProxmoxService.getVMStatus] Failed to get VM status: %s", body)
		return "", fmt.Errorf("failed to get VM status: %s", body)
	}

	var vmStatus VMStatusResponse
	if err := json.Unmarshal(body, &vmStatus); err != nil {
		log.Errorf("[ProxmoxService.getVMStatus] Failed to unmarshal response: %v", err)
		return "", err
	}

	log.Debug("[ProxmoxService.getVMStatus] VM status [%s]: %s", user.ID, vmStatus.Data.Status)

	return vmStatus.Data.Status, nil
}

type VMInfo struct {
	VMID   int    `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type VMInfoResponse struct {
	Data []VMInfo `json:"data"`
}

func (p *ProxmoxService) vmExists(id int) (bool, error) {
	ret := false
	url := fmt.Sprintf("%s/nodes/%s/qemu", p.hostUrl, p.Node, user.ID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("[ProxmoxService.vmExists] Failed to create request: %v", err)
		return ret, err
	}

	err = p.getTicket()
	if err != nil {
		log.Errorf("[ProxmoxService.vmExists] Failed to get ticket: %v", err)
		return false, err
	}

	req.Header.Set("CSRFPreventionToken", p.CSRFPreventionToken)
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", p.Ticket))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("[ProxmoxService.vmExists] Failed to make request: %v", err)
		return ret, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[ProxmoxService.vmExists] Failed to read response body: %v", err)
		return ret, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("[ProxmoxService.vmExists] Failed to get VM status: %s", body)
		return ret, fmt.Errorf("failed to check if VM exists: %s", body)
	}

	var vmInfos VMInfoResponse
	if err := json.Unmarshal(body, &vmInfos); err != nil {
		log.Errorf("[ProxmoxService.vmExists] Failed to unmarshal response: %v", err)
		return ret, err
	}

	for _, vmInfo := range vmInfos.Data {
		if vmInfo.VMID == id {
			log.Info("[ProxmoxService.vmExists] VM exists: %d", id)
			ret = true
			break
		}
	}

	if !ret {
		log.Infof("[ProxmoxService.vmExists] VM does not exist: %d", id)
	}

	return ret, nil
}

func (p *ProxmoxService) Run(user *domain.User) error {
	exists, err := p.vmExists(user.ID)

	if err != nil {
		log.Errorf("[ProxmoxService.Run] Failed to check if VM exists: %v", err)
		return err
	}

	if !exists {
		log.Info("[ProxmoxService.Run] VM does not exist: %d", user.ID)
		err := p.createVM(user)
		if err != nil {
			log.Errorf("[ProxmoxService.Run] Failed to create VM: %v", err)
			return err
		}
	}

	vmStatus, err := p.getVMStatus(user)

	if err != nil {
		log.Errorf("[ProxmoxService.Run] Failed to get VM status: %v", err)
		return err
	}

	if vmStatus == "running" {
		log.Infof("[ProxmoxService.Run] VM is already running: %d", user.ID)
		return nil
	}

	log.Infof("[ProxmoxService.Run] Starting VM: %d", user.ID)
	url := fmt.Sprintf("%s/nodes/%s/qemu/%d/status/start", p.hostUrl, p.Node, user.ID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Errorf("[ProxmoxService.Run] Failed to create request: %v", err)
		return err
	}

	err = p.getTicket()
	if err != nil {
		log.Errorf("[ProxmoxService.Run] Failed to get ticket: %v", err)
		return err
	}

	req.Header.Set("CSRFPreventionToken", p.CSRFPreventionToken)
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", p.Ticket))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("[ProxmoxService.Run] Failed to make request: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[ProxmoxService.Run] Failed to read response body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("[ProxmoxService.Run] Failed to get VM status: %s", body)
		return fmt.Errorf("failed to start VM: %s", body)
	}

	log.Infof("[ProxmoxService.Run] VM started successfully: %d", user.ID)

	return nil
}

type CloneData struct {
	NewID  int    `json:"newid"`
	Name   string `json:"name"`
	Full   bool   `json:"full"`
	Target string `json:"target,omitempty"`
}

func (p *ProxmoxService) createVM(user *domain.User) error {
	url := fmt.Sprintf("%s/nodes/%s/qemu/%d/clone", p.hostUrl, p.Node, p.TemplateID)

	data := CloneData{
		NewID: user.ID,
		Name:  fmt.Sprintf("server-%s", user.Login),
		Full:  true, // Defina como `true` para clonar o disco completamente
	}

	// Converte os dados para JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("erro ao converter os dados para JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("[ProxmoxService.createVM] Failed to create request: %v", err)
		return err
	}

	err = p.getTicket()
	if err != nil {
		log.Errorf("[ProxmoxService.createVM] Failed to get ticket: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("CSRFPreventionToken", p.CSRFPreventionToken)
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", p.Ticket))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("[ProxmoxService.createVM] Failed to make request: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[ProxmoxService.createVM] Failed to read response body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("[ProxmoxService.createVM] Failed to get VM status: %s", body)
		return fmt.Errorf("failed to try to create VM: %s", body)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Errorf("[ProxmoxService.createVM] Failed to decode response: %v", err)
		return fmt.Errorf("erro ao decodificar a resposta: %v", err)
	}

	log.Infof("[ProxmoxService.createVM] Request to create a VM successfully: %s[%d] -> %s", user.Login, user.ID, result["data"])

	return nil
}

func (p *ProxmoxService) Stop(user *domain.User) error {
	url := fmt.Sprintf("%s/nodes/%s/qemu/%d/status/stop", p.hostUrl, p.Node, user.ID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Errorf("[ProxmoxService.Stop] Failed to create request: %v", err)
		return err
	}

	err = p.getTicket()
	if err != nil {
		log.Errorf("[ProxmoxService.Stop] Failed to get ticket: %v", err)
		return err
	}

	req.Header.Set("CSRFPreventionToken", p.CSRFPreventionToken)
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", p.Ticket))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("[ProxmoxService.Stop] Failed to make request: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[ProxmoxService.Stop] Failed to read response body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("[ProxmoxService.Stop] Failed to get VM status: %s", body)
		return fmt.Errorf("failed to execute stop command: %s", body)
	}

	log.Infof("[ProxmoxService.Stop] VM stopped successfully: %d", user.ID)

	return nil
}

type ExecData struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func (p *ProxmoxService) runBootstrapScript(user *domain.User) error {
	url := fmt.Sprintf("%s/nodes/%s/qemu/%d/agent/exec", p.hostUrl, p.Node, user.ID)
	script := "#!/bin/bash\n" +
		"echo 'Running bootstrap script...'\n" +
		"echo 'Bootstrap script completed.'\n"

	data := ExecData{
		Command: "bash",
		Args:    []string{"-c", script},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Errorf("[ProxmoxService.runBootstrapScript] Failed to marshal JSON: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("[ProxmoxService.runBootstrapScript] Failed to create request: %v", err)
		return err
	}

	err = p.getTicket()
	if err != nil {
		log.Errorf("[ProxmoxService.runBootstrapScript] Failed to get ticket: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("CSRFPreventionToken", p.CSRFPreventionToken)
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", p.Ticket))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("[ProxmoxService.runBootstrapScript] Failed to make request: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[ProxmoxService.runBootstrapScript] Failed to read response body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("[ProxmoxService.runBootstrapScript] Failed to get VM status: %s", body)
		return fmt.Errorf("failed to execute bootstrap script: %s", body)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Errorf("[ProxmoxService.runBootstrapScript] Failed to decode response: %v", err)
		return fmt.Errorf("erro ao decodificar a resposta: %v", err)
	}

	status, ok := result["data"].(map[string]interface{})["status"].(string)
	if !ok {
		log.Errorf("[ProxmoxService.runBootstrapScript] Failed to get status from response: %v", result)
		return fmt.Errorf("status não encontrado na resposta")
	}

	log.Infof("[ProxmoxService.runBootstrapScript] Bootstrap script executed successfully: %s", status)

	return nil
}
