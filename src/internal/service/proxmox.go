package service

import (
	"code-server-launcher/internal/config"
	"code-server-launcher/internal/domain"
	"code-server-launcher/internal/logger"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

type ProxmoxService struct {
	config.ProxmoxConfig
	log           *logger.Logger
	apiURL        string
	proxmoxClient *proxmox.Client
}

func NewLXCService(cfg *config.ProxmoxConfig) *ProxmoxService {
	ret := &ProxmoxService{
		log:           logger.NewLogger("ProxmoxService"),
		apiURL:        "https://" + cfg.Host + "/api2/json",
		ProxmoxConfig: *cfg,
	}

	ctx := context.Background()

	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	var err error

	ret.proxmoxClient, err = proxmox.NewClient(
		ret.apiURL,
		httpClient,
		"",
		tlsConfig,
		"",
		300,
	)

	if err != nil {
		ret.log.Error("Failed to create Proxmox client: %v -> config: %v", err, cfg)
		return nil
	}

	err = ret.proxmoxClient.Login(ctx, cfg.Username, cfg.Password, "")
	if err != nil {
		ret.log.Error("Failed to login to Proxmox: %v", err)
	} else {
		ret.log.Info("Proxmox client created successfully")
	}

	return ret
}

func (p *ProxmoxService) Run(user *domain.User) error {
	p.log.Info("Running LXC for user: %d", user.ID)

	status, err := p.getStatus(user)

	if err != nil {
		p.log.Error("Failed to get LXC status: %v", err)
		return err
	}

	if status == domain.VmStatusRunning {
		p.log.Info("LXC container already exists for user: %d", user.ID)
		return nil
	}

	if status == domain.VmStatusStopped || status == domain.VmStatusPaused || status == domain.VmStatusSuspended {
		p.log.Info("LXC container already exists for user %d and is stopped", user.ID)
		p.log.Info("Starting LXC container for user: %d", user.ID)
		err := p.TurnOnContainer(user)
		if err != nil {
			p.log.Error("Failed to start LXC container: %v", err)
			return err
		}
		return nil
	}

	err = p.CreateContainer(user)

	if err != nil {
		p.log.Error("Failed to create LXC container: %v", err)
		return err
	}

	p.log.Info("LXC container created successfully for user: %d", user.ID)

	err = p.TurnOnContainer(user)
	if err != nil {
		p.log.Error("Failed to start LXC container: %v", err)
		return err
	}

	p.log.Info("LXC container started successfully for user: %d", user.ID)

	return nil
}

func (p *ProxmoxService) Stop(user *domain.User) error {
	p.log.Info("Stopping LXC for user: %d", user.ID)
	return nil
}

func (p *ProxmoxService) Exists(user *domain.User) (bool, error) {
	p.log.Debug("Checking if LXC exists for user: %d", user.ID)

	info, err := p.GetInfo(user)

	if err != nil || info == nil {
		p.log.Info("LXC does not exist for user: %d -> %S", user.ID, err)
		return false, err
	}

	return true, nil
}

func (p *ProxmoxService) getStatus(user *domain.User) (domain.VmStatus, error) {
	p.log.Debug("Checking if LXC is running for user: %d", user.ID)
	info, err := p.GetInfo(user)

	if err != nil {
		p.log.Error("Failed to get LXC info: %v", err)
		return domain.VmStatusUnknown, err
	}

	if info == nil {
		p.log.Info("LXC does not exist for user: %d", user.ID)
		return domain.VmStatusMissing, nil
	}

	return info.Status, nil
}

func (p *ProxmoxService) GetInfo(user *domain.User) (*domain.VmInfo, error) {
	p.log.Debug("Checking status of LXC for user: %d", user.ID)

	ctx := context.Background()

	vmRef := proxmox.NewVmRef(proxmox.GuestID(user.ID))

	ret, err := p.proxmoxClient.GetVmInfo(ctx, vmRef)

	if err != nil {
		p.log.Error("Failed to get VM list: %v", err)
		return nil, err
	}

	vm, err := domain.ParseVmInfo(ret)

	if err != nil {
		p.log.Error("Failed to parse VM list: %v", err)
		return nil, err
	}

	return vm, nil
}

func (p *ProxmoxService) CreateContainer(user *domain.User) error {
	p.log.Info("Creating LXC container for user: %d", user.ID)

	ctx := context.Background()

	templateRef := proxmox.NewVmRef(proxmox.GuestID(p.TemplateID))
	templateRef.SetNode(p.Node)

	newContainerName := fmt.Sprintf("codeserver-%s", user.Login)
	guestId := (proxmox.GuestID)(user.ID)

	target := proxmox.CloneLxcTarget{
		Full: &proxmox.CloneLxcFull{
			Node:    (proxmox.NodeName)(p.Node),
			ID:      &guestId,
			Name:    &newContainerName,
			Storage: &p.StorageName,
		},
	}

	targetRef, err := templateRef.CloneLxc(ctx, target, p.proxmoxClient)

	if err != nil {
		p.log.Error("Failed to clone LXC container: %v", err)
		return err
	}

	if targetRef == nil {
		p.log.Error("Failed to clone LXC container: targetRef is nil")
		return fmt.Errorf("failed to clone LXC container: targetRef is nil")
	}

	cfg, err := proxmox.NewConfigLxcFromApi(ctx, targetRef, p.proxmoxClient)
	if err != nil {
		p.log.Error("Failed to get LXC config: %v", err)
		return err
	}

	cfg.Memory = p.MemSize
	cfg.Cores = p.CPUCores
	cfg.Networks = proxmox.QemuDevices{
		0: {
			"name":     "eth0",
			"bridge":   p.NetworkInterface,
			"firewall": true,
			"ip":       fmt.Sprintf(p.BaseIP, user.ID),
		},
	}

	err = cfg.UpdateConfig(ctx, targetRef, p.proxmoxClient)

	if err != nil {
		p.log.Error("Failed to update LXC config: %v", err)
		return err
	}

	return nil
}

func (p *ProxmoxService) HibernateVm(user *domain.User) error {
	p.log.Info("Hibernating LXC container for user: %d", user.ID)

	ctx := context.Background()
	vmRef := proxmox.NewVmRef(proxmox.GuestID(user.ID))

	status, err := p.proxmoxClient.HibernateVm(ctx, vmRef)

	if err != nil {
		p.log.Error("Failed to hibernate LXC container: %v", err)
		return err
	}

	p.log.Info("LXC container hibernated successfully for user: %d -> Status: %s", user.ID, status)

	return nil
}

func (p *ProxmoxService) StopContainer(user *domain.User) error {
	p.log.Info("Stopping LXC container for user: %d", user.ID)

	ctx := context.Background()
	vmRef := proxmox.NewVmRef(proxmox.GuestID(user.ID))

	status, err := p.proxmoxClient.StopVm(ctx, vmRef)

	if err != nil {
		p.log.Error("Failed to stop LXC container: %v", err)
		return err
	}

	p.log.Info("LXC container stopped successfully for user: %d -> Status: %s", user.ID, status)

	return nil
}

func (p *ProxmoxService) TurnOnContainer(user *domain.User) error {
	p.log.Info("Turning on LXC container for user: %d", user.ID)

	ctx := context.Background()
	vmRef := proxmox.NewVmRef(proxmox.GuestID(user.ID))

	status, err := p.proxmoxClient.StartVm(ctx, vmRef)

	if err != nil {
		p.log.Error("Failed to start LXC container: %v", err)
		return err
	}

	p.log.Info("LXC container started successfully for user: %d -> Status: %s", user.ID, status)

	time.Sleep(5 * time.Second)
	for i := 0; i < p.TimetoStart; i++ {
		time.Sleep(time.Second)
		info, err := p.GetInfo(user)
		if err != nil {
			p.log.Error("Failed to get LXC info: %v", err)
			return err
		}
		if info.Status == domain.VmStatusRunning {
			p.log.Info("LXC container is running for user: %d", user.ID)
			return nil
		}

		p.log.Info("Waiting for LXC container to start for user: %d -> Status: %s", user.ID, info.Status)
	}

	return fmt.Errorf("LXC container did not start in time for user: %d", user.ID)
}
