package tools

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// SSHConfig holds SSH connection parameters
type SSHConfig struct {
	Host   string
	Port   int
	User   string
	Pass   string
	Key    string
	Client *ssh.Client
}

// NewSSHConfig creates a new SSH configuration
func NewSSHConfig(host string, port int, user, pass, key string) *SSHConfig {
	return &SSHConfig{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		Key:  key,
	}
}

// Connect establishes an SSH connection
func (s *SSHConfig) Connect() error {
	var auth []ssh.AuthMethod

	if s.Key != "" {
		key, err := os.ReadFile(s.Key)
		if err != nil {
			return err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	if s.Pass != "" {
		auth = append(auth, ssh.Password(s.Pass))
	}

	config := &ssh.ClientConfig{
		User:            s.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Insecure for testing, use proper host key checking in production
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port), config)
	if err != nil {
		return err
	}

	s.Client = client
	return nil
}

// Execute runs a command on the remote host
func (s *SSHConfig) Execute(cmd string) (string, error) {
	if s.Client == nil {
		if err := s.Connect(); err != nil {
			return "", err
		}
	}

	session, err := s.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// Close closes the SSH connection
func (s *SSHConfig) Close() {
	if s.Client != nil {
		s.Client.Close()
	}
}
