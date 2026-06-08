package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/luncher4/vpn-panel/internal/models"
)

const xrayConfigPath = "/usr/local/etc/xray/config.json"

type XrayManager struct{}

func NewXrayManager() *XrayManager { return &XrayManager{} }

func (m *XrayManager) AddInbounds(inbounds []models.Inbound) error {
	cfg, err := readXrayConfig()
	if err != nil {
		return fmt.Errorf("read config: %v", err)
	}
	for _, ib := range inbounds {
		cfg.Inbounds = append(cfg.Inbounds, buildInboundJSON(ib))
	}
	if err := writeXrayConfig(cfg); err != nil {
		return fmt.Errorf("write config: %v", err)
	}
	return restartXray()
}

func (m *XrayManager) RemoveInbound(port int) error {
	cfg, err := readXrayConfig()
	if err != nil {
		return fmt.Errorf("read config: %v", err)
	}
	var newInbounds []json.RawMessage
	for _, ib := range cfg.Inbounds {
		var parsed map[string]interface{}
		json.Unmarshal(ib, &parsed)
		p, _ := parsed["port"].(float64)
		if int(p) != port {
			newInbounds = append(newInbounds, ib)
		}
	}
	cfg.Inbounds = newInbounds
	if err := writeXrayConfig(cfg); err != nil {
		return fmt.Errorf("write config: %v", err)
	}
	return restartXray()
}

type xrayConfigFile struct {
	Log       interface{}             `json:"log"`
	API       interface{}             `json:"api,omitempty"`
	Inbounds  []json.RawMessage       `json:"inbounds"`
	Outbounds []json.RawMessage       `json:"outbounds"`
	Routing   interface{}             `json:"routing,omitempty"`
}

func readXrayConfig() (*xrayConfigFile, error) {
	data, err := os.ReadFile(xrayConfigPath)
	if err != nil {
		return nil, err
	}
	cfg := &xrayConfigFile{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func writeXrayConfig(cfg *xrayConfigFile) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(xrayConfigPath, data, 0644)
}

func restartXray() error {
	cmd := exec.Command("systemctl", "restart", "xray")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restart xray failed: %s: %v", string(out), err)
	}
	time.Sleep(1 * time.Second)
	return nil
}

func buildInboundJSON(ib models.Inbound) json.RawMessage {
	streamSettings := map[string]interface{}{
		"network":  ib.Transport,
		"security": "reality",
		"realitySettings": map[string]interface{}{
			"dest":        "www.microsoft.com:443",
			"shortIds":    []string{"19e72187"},
			"serverNames": []string{"www.microsoft.com", "www.bing.com", "www.cloudflare.com", "www.github.com"},
			"privateKey":  "eJvaNmNJP-X6l2MprNA8KWPGV0aRgG_er6jArLrKZmY",
		},
	}

	if ib.Settings != "" && ib.Settings != "{}" {
		var extra map[string]interface{}
		if json.Unmarshal([]byte(ib.Settings), &extra) == nil {
			for k, v := range extra {
				streamSettings[k] = v
			}
		}
	}

	inbound := map[string]interface{}{
		"port":     ib.Port,
		"protocol": ib.Protocol,
		"settings": map[string]interface{}{
			"clients": []map[string]interface{}{
				{"id": "417b6335-74f0-4597-8e23-4ab8ee6a3ced", "flow": "xtls-rprx-vision"},
			},
			"decryption": "none",
		},
		"streamSettings": streamSettings,
		"tag":            fmt.Sprintf("%s-%d", ib.Protocol, ib.Port),
	}

	data, _ := json.Marshal(inbound)
	return data
}


func (m *XrayManager) AddClientToInbound(port int, clientUUID string) error {
	cfg, err := readXrayConfig()
	if err != nil {
		return fmt.Errorf("read config: %v", err)
	}
	for i, ib := range cfg.Inbounds {
		var parsed map[string]interface{}
		json.Unmarshal(ib, &parsed)
		p, _ := parsed["port"].(float64)
		if int(p) == port {
			// Добавляем клиента в inbound
			clients, _ := parsed["settings"].(map[string]interface{})["clients"].([]interface{})
			clients = append(clients, map[string]interface{}{
				"id":   clientUUID,
				"flow": "xtls-rprx-vision",
			})
			if settings, ok := parsed["settings"].(map[string]interface{}); ok {
				settings["clients"] = clients
				parsed["settings"] = settings
			}
			data, _ := json.Marshal(parsed)
			cfg.Inbounds[i] = data
			break
		}
	}
	if err := writeXrayConfig(cfg); err != nil {
		return fmt.Errorf("write config: %v", err)
	}
	return restartXray()
}
func GetPublicKey() string {
	return "YI3mGqA_bSX9tIPEB5RL2pPxR8bmC1L_GQqz4v5CxlA"
}
