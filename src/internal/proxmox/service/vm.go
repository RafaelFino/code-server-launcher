/*package service

import (
	"bytes"
	"code-server-launcher/internal/code-server-laucher/domain"
	"code-server-launcher/internal/code-server-laucher/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type VM struct {
	log        *logger.Logger
	auth       *Auth
	host       string
	node       string
	templateID int
}

func NewVMService(config *domain.ProxmoxConfig, auth *Auth) *VM {
	return &VM{
		log:        logger.NewLogger("ProxmoxVMService"),
		auth:       auth,
		host:       config.Host,
		node:       config.Node,
		templateID: config.TemplateID,
	}
}

//	ret.hostUrl = "https://" + ret.Host + "/api2/json"

func (v *VM) Run(user *domain.User) error {
	exists, err := v.vmExists(user.ID)

	if err != nil {
		v.log.Error("Failed to check if VM exists: %v", err)
		return err
	}

	if !exists {
		v.log.Info("VM does not exist: %d", user.ID)
		err := v.createVM(user)
		if err != nil {
			v.log.Error("Failed to create VM: %v", err)
			return err
		}
	}

	vmStatus, err := v.getStatus(user)

	if err != nil {
		v.log.Error("Failed to get VM status: %v", err)
		return err
	}

	if vmStatus == "running" {
		v.log.Info("VM is already running: %d", user.ID)
		return nil
	}

	v.log.Info("Starting VM: %d", user.ID)
	url := fmt.Sprintf("%s/nodes/%s/qemu/%d/status/start", v.host, v.node, user.ID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		v.log.Error("Failed to create request: %v", err)
		return err
	}

	err = v.auth.GetSession(&req.Header)
	if err != nil {
		v.log.Error("Failed to get ticket: %v", err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		v.log.Error("Failed to make request: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		v.log.Error("Failed to read response body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		v.log.Error("Failed to get VM status: %s", body)
		return fmt.Errorf("failed to start VM: %s", body)
	}

	v.log.Info("VM started successfully: %d", user.ID)

	return nil
}

func (v *VM) Stop(user *domain.User) error {
	url := fmt.Sprintf("%s/nodes/%s/qemu/%d/status/stop", v.host, v.node, user.ID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		v.log.Error("Failed to create request: %v", err)
		return err
	}

	err = v.auth.GetSession(&req.Header)
	if err != nil {
		v.log.Error("Failed to get ticket: %v", err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		v.log.Error("Failed to make request: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		v.log.Error("Failed to read response body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		v.log.Error("Failed to get VM status: %s", body)
		return fmt.Errorf("failed to execute stop command: %s", body)
	}

	v.log.Info("VM stopped successfully: %d", user.ID)

	return nil
}

func (v *VM) getStatus(user *domain.User) (string, error) {
	url := fmt.Sprintf("%s/nodes/%s/qemu/%s/status/current", v.host, v.node, user.ID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		v.log.Error("Failed to create request: %v", err)
		return "", err
	}

	err = v.auth.GetSession(&req.Header)
	if err != nil {
		v.log.Error("Failed to get ticket: %v", err)
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		v.log.Error("Failed to make request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		v.log.Error("Failed to read response body: %v", err)
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		v.log.Error("Failed to get VM status: %s", body)
		return "", fmt.Errorf("failed to get VM status: %s", body)
	}

	var vmStatus domain.VMStatusResponse
	if err := json.Unmarshal(body, &vmStatus); err != nil {
		v.log.Error("Failed to unmarshal response: %v", err)
		return "", err
	}

	v.log.Debug("VM status [%s]: %s", user.ID, vmStatus.Data.Status)

	return vmStatus.Data.Status, nil
}

func (v *VM) vmExists(id int) (bool, error) {
	ret := false
	url := fmt.Sprintf("%s/nodes/%s/qemu", v.host, v.node, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		v.log.Error("Failed to create request: %v", err)
		return ret, err
	}

	err = v.auth.GetSession(&req.Header)
	if err != nil {
		v.log.Error("Failed to get ticket: %v", err)
		return false, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		v.log.Error("Failed to make request: %v", err)
		return ret, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		v.log.Error("Failed to read response body: %v", err)
		return ret, err
	}

	if resp.StatusCode != http.StatusOK {
		v.log.Error("Failed to get VM status: %s", body)
		return ret, fmt.Errorf("failed to check if VM exists: %s", body)
	}

	var vmInfos domain.VMInfoResponse
	if err := json.Unmarshal(body, &vmInfos); err != nil {
		v.log.Error("Failed to unmarshal response: %v", err)
		return ret, err
	}

	for _, vmInfo := range vmInfos.Data {
		if vmInfo.VMID == id {
			v.log.Info("VM exists: %d", id)
			ret = true
			break
		}
	}

	if !ret {
		v.log.Info("VM does not exist: %d", id)
	}

	return ret, nil
}

func (v *VM) createVM(user *domain.User) error {
	url := fmt.Sprintf("%s/nodes/%s/qemu/%d/clone", v.host, v.node, v.templateID)

	data := domain.CloneData{
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
		v.log.Error("Failed to create request: %v", err)
		return err
	}

	err = v.auth.GetSession(&req.Header)
	if err != nil {
		v.log.Error("Failed to get ticket: %v", err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		v.log.Error("Failed to make request: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		v.log.Error("Failed to read response body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		v.log.Error("Failed to get VM status: %s", body)
		return fmt.Errorf("failed to try to create VM: %s", body)
	}

	var result domain.ProxmoxResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		v.log.Error("Failed to decode response: %v", err)
		return fmt.Errorf("erro ao decodificar a resposta: %v", err)
	}

	v.log.Info("-> %s", user.Login, user.ID, result)

	return nil
}

func (v *VM) runBootstrapScript(user *domain.User) error {
	url := fmt.Sprintf("%s/nodes/%s/qemu/%d/agent/exec", v.host, v.node, user.ID)
	script := "#!/bin/bash\n" +
		"echo 'Running bootstrap script...'\n" +
		"echo 'Bootstrap script completed.'\n"

	data := domain.ExecData{
		Command: "bash",
		Args:    []string{"-c", script},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		v.log.Error("Failed to marshal JSON: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		v.log.Error("Failed to create request: %v", err)
		return err
	}

	err = v.auth.GetSession(&req.Header)
	if err != nil {
		v.log.Error("Failed to get ticket: %v", err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		v.log.Error("Failed to make request: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		v.log.Error("Failed to read response body: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		v.log.Error("Failed to get VM status: %s", body)
		return fmt.Errorf("failed to execute bootstrap script: %s", body)
	}

	var result domain.ProxmoxResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		v.log.Error("Failed to decode response: %v", err)
		return fmt.Errorf("erro ao decodificar a resposta: %v", err)
	}

	status, ok := result.Data["status"].(string)
	if !ok {
		v.log.Error("Failed to get status from response: %v", result)
		return fmt.Errorf("status não encontrado na resposta")
	}

	v.log.Info("Bootstrap script executed successfully: %s", status)

	return nil
}
*/