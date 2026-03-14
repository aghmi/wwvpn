package node

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type XRayConfig struct {
	ConfigPath string
	PublicKey  string
	ShortID   string
	SNI       string
	Port      int
}

type XRayManager struct {
	config XRayConfig
	mu     sync.Mutex
	users  map[string]string
}

func NewXRayManager(config XRayConfig) *XRayManager {
	return &XRayManager{
		config: config,
		users:  make(map[string]string),
	}
}

func (m *XRayManager) AddUser(deviceUUID string) (vlessUUID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.users[deviceUUID]; ok {
		return existing, nil
	}

	vlessUUID = uuid.New().String()

	cmd := exec.Command("xray", "api", "adi",
		"--server=127.0.0.1:10085",
		"-tag=vless-in",
		fmt.Sprintf(`-id=%s`, vlessUUID),
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("xray api add user failed: %s: %w", string(output), err)
	}

	m.users[deviceUUID] = vlessUUID
	return vlessUUID, nil
}

func (m *XRayManager) RemoveUser(vlessUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := exec.Command("xray", "api", "rmi",
		"--server=127.0.0.1:10085",
		"-tag=vless-in",
		fmt.Sprintf(`-id=%s`, vlessUUID),
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("xray api remove user failed: %s: %w", string(output), err)
	}

	for k, v := range m.users {
		if v == vlessUUID {
			delete(m.users, k)
			break
		}
	}

	return nil
}

func (m *XRayManager) UserCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.users)
}

func (m *XRayManager) Params() (publicKey, shortID, sni string, port int) {
	return m.config.PublicKey, m.config.ShortID, m.config.SNI, m.config.Port
}

func GenerateXRayKeys() (privateKey, publicKey string, err error) {
	cmd := exec.Command("xray", "x25519")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("xray x25519 failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Private key:") {
			privateKey = strings.TrimSpace(strings.TrimPrefix(line, "Private key:"))
		}
		if strings.HasPrefix(line, "Public key:") {
			publicKey = strings.TrimSpace(strings.TrimPrefix(line, "Public key:"))
		}
	}

	if privateKey == "" || publicKey == "" {
		return "", "", fmt.Errorf("failed to parse xray x25519 output")
	}

	return privateKey, publicKey, nil
}

func GenerateXRayShortID() string {
	cmd := exec.Command("openssl", "rand", "-hex", "8")
	output, err := cmd.Output()
	if err != nil {
		return "abcdef1234567890"
	}
	return strings.TrimSpace(string(output))
}

func WriteXRayConfig(path, privateKey, shortID, sni string, port int) error {
	config := map[string]interface{}{
		"log": map[string]string{
			"loglevel": "warning",
		},
		"api": map[string]interface{}{
			"tag":      "api",
			"services": []string{"HandlerService"},
		},
		"inbounds": []map[string]interface{}{
			{
				"tag":      "api-in",
				"listen":   "127.0.0.1",
				"port":     10085,
				"protocol": "dokodemo-door",
				"settings": map[string]string{
					"address": "127.0.0.1",
				},
			},
			{
				"tag":      "vless-in",
				"listen":   "0.0.0.0",
				"port":     port,
				"protocol": "vless",
				"settings": map[string]interface{}{
					"clients":    []interface{}{},
					"decryption": "none",
				},
				"streamSettings": map[string]interface{}{
					"network":  "tcp",
					"security": "reality",
					"realitySettings": map[string]interface{}{
						"dest":        sni + ":443",
						"serverNames": []string{sni},
						"privateKey":  privateKey,
						"shortIds":    []string{shortID},
					},
				},
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"protocol": "freedom",
				"tag":      "direct",
			},
		},
		"routing": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"inboundTag":  []string{"api-in"},
					"outboundTag": "api",
					"type":        "field",
				},
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
