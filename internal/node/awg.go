package node

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
)

type AWGConfig struct {
	Interface  string
	ListenPort int
	PrivateKey string
	PublicKey  string
	Subnet    string
	Jc        int
	Jmin      int
	Jmax      int
	S1        int
	S2        int
	H1        int
	H2        int
	H3        int
	H4        int
}

type AWGManager struct {
	config    AWGConfig
	mu        sync.Mutex
	nextIP    byte
	allocated map[string]string
}

func NewAWGManager(config AWGConfig) *AWGManager {
	return &AWGManager{
		config:    config,
		nextIP:    2,
		allocated: make(map[string]string),
	}
}

func (m *AWGManager) AddPeer(clientPubKey string) (assignedIP string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ip, exists := m.allocated[clientPubKey]; exists {
		return ip, nil
	}

	assignedIP = fmt.Sprintf("10.0.0.%d", m.nextIP)
	m.nextIP++

	cmd := exec.Command("awg", "set", m.config.Interface,
		"peer", clientPubKey,
		"allowed-ips", assignedIP+"/32",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("wg set failed: %s: %w", string(output), err)
	}

	m.allocated[clientPubKey] = assignedIP
	return assignedIP, nil
}

func (m *AWGManager) RemovePeer(clientPubKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := exec.Command("awg", "set", m.config.Interface,
		"peer", clientPubKey, "remove",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wg remove failed: %s: %w", string(output), err)
	}

	delete(m.allocated, clientPubKey)
	return nil
}

func (m *AWGManager) PeerCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.allocated)
}

func (m *AWGManager) PublicKey() string {
	return m.config.PublicKey
}

func (m *AWGManager) Endpoint() string {
	return fmt.Sprintf("%s:%d", getPublicIP(), m.config.ListenPort)
}

func (m *AWGManager) ObfuscationParams() (jc, jmin, jmax, s1, s2, h1, h2, h3, h4 int) {
	c := m.config
	return c.Jc, c.Jmin, c.Jmax, c.S1, c.S2, c.H1, c.H2, c.H3, c.H4
}

func getPublicIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "0.0.0.0"
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}

func GenerateAWGKeyPair() (privateKey, publicKey string, err error) {
	privCmd := exec.Command("awg", "genkey")
	privOut, err := privCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("wg genkey failed: %w", err)
	}
	privateKey = strings.TrimSpace(string(privOut))

	pubCmd := exec.Command("awg", "pubkey")
	pubCmd.Stdin = strings.NewReader(privateKey)
	pubOut, err := pubCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("wg pubkey failed: %w", err)
	}
	publicKey = strings.TrimSpace(string(pubOut))

	return privateKey, publicKey, nil
}
