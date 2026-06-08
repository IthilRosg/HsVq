package services

import (
	"encoding/json"
	"fmt"
)

// ConfigTemplate — шаблон конфигурации inbound
type ConfigTemplate struct {
	Name        string   `json:"name"`
	Protocol    string   `json:"protocol"`
	Transport   string   `json:"transport"`
	Port        int      `json:"port"`
	Security    string   `json:"security"`
	Description string   `json:"description"`
	DefaultSNI  string   `json:"default_sni"`
	Fingerprint string   `json:"fingerprint"`
	ServerNames []string `json:"server_names"`
	PortNote    string   `json:"port_note"`
}

// GetConfigTemplates возвращает список шаблонов для inbound
func GetConfigTemplates() []ConfigTemplate {
	return []ConfigTemplate{
		{
			Name:        "REALITY + Vision (рекомендуется)",
			Protocol:    "vless",
			Transport:   "tcp",
			Port:        443,
			Security:    "reality",
			Description: "Базовый REALITY с Vision flow. Лучший баланс скорости и скрытности",
			DefaultSNI:  "www.cloudflare.com",
			Fingerprint: "chrome",
			ServerNames: []string{"www.cloudflare.com", "www.github.com", "www.digitalocean.com"},
			PortNote:    "443 — стандартный HTTPS, не блокируется DPI",
		},
		{
			Name:        "REALITY + H2 (для фильмов)",
			Protocol:    "vless",
			Transport:   "h2",
			Port:        443,
			Security:    "reality",
			Description: "Мультиплексирование H2. Лучше для стриминга и скачивания",
			DefaultSNI:  "www.cloudflare.com",
			Fingerprint: "chrome",
			ServerNames: []string{"www.cloudflare.com", "www.github.com", "www.digitalocean.com"},
			PortNote:    "443 — мультиплексирование, выше скорость на больших файлах",
		},
		{
			Name:        "REALITY + ECH + fragment (MaxProbiv)",
			Protocol:    "vless",
			Transport:   "tcp",
			Port:        443,
			Security:    "reality",
			Description: "Максимальная защита: шифрование SNI (ECH) + фрагментация",
			DefaultSNI:  "www.cloudflare.com",
			Fingerprint: "chrome",
			ServerNames: []string{"www.cloudflare.com", "www.github.com"},
			PortNote:    "443 — максимальная защита от DPI",
		},
		{
			Name:        "Hysteria 2 (резервный протокол)",
			Protocol:    "hysteria2",
			Transport:   "quic",
			Port:        8443,
			Security:    "none",
			Description: "Hysteria 2 на QUIC. Хорош как запасной, если REALITY заблокирован",
			DefaultSNI:  "",
			Fingerprint: "",
			ServerNames: []string{},
			PortNote:    "8443 — альтернативный порт, QUIC протокол",
		},
	}
}

// ValidateInboundConfig проверяет конфиг inbound перед сохранением
func ValidateInboundConfig(protocol string, port int, security string, transport string) error {
	if protocol == "" {
		return fmt.Errorf("protocol is required")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if port < 1024 && port != 443 && port != 8443 {
		return fmt.Errorf("port %d is privileged (use 443 or >1024)", port)
	}
	if security == "reality" && protocol != "vless" {
		return fmt.Errorf("REALITY security is only supported with VLESS protocol")
	}
	validProtocols := map[string]bool{"vless": true, "shadowsocks": true, "trojan": true, "hysteria2": true}
	if !validProtocols[protocol] {
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
	validTransports := map[string]bool{"tcp": true, "ws": true, "grpc": true, "h2": true, "quic": true}
	if !validTransports[transport] {
		return fmt.Errorf("unsupported transport: %s", transport)
	}
	return nil
}

// ApplyTemplateToSettings применяет шаблон к настройкам inbound
func ApplyTemplateToSettings(template ConfigTemplate) string {
	settings := map[string]interface{}{
		"fingerprint":  template.Fingerprint,
		"serverName":   template.DefaultSNI,
		"serverNames":  template.ServerNames,
		"shortIds":     []string{"19e72187"},
		"privateKey":   "eJvaNmNJP-X6l2MprNA8KWPGV0aRgG_er6jArLrKZmY",
	}
	if template.Protocol == "hysteria2" {
		settings = map[string]interface{}{
			"up_mbps":   100,
			"down_mbps": 500,
			"obfs":      "salamander",
		}
	}
	data, _ := json.Marshal(settings)
	return string(data)
}
