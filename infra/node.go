package infra

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/metrue/go-ssh-client"
	"github.com/mitchellh/go-homedir"
)

const nodeTypeMaster = "k3s_master"
const nodeTypeAgent = "k3s_agent"
const nodeTypeDocker = "docker_agent"

// Noder node interface
type Noder interface {
	Provision(meta map[string]string) error
	GetType() string
	GetName() string
	GetToken() (string, error)
	GetIP() string
	Dump() map[string]string
}

// Node define a node
type Node struct {
	IP   string `json:"ip"`
	User string `json:"user"`

	Type string `json:"type"`
	Name string `json:"name"`

	sshClient ssh.Clienter
}

func createNode(meta map[string]string) (Noder, error) {
	var ip string
	var user string
	var typ string
	var name string
	for attr, value := range meta {
		if attr == "type" {
			typ = value
		}
		if attr == "name" {
			name = value
		}
		if attr == "ip" {
			ip = value
		}
		if attr == "user" {
			user = value
		}
	}
	return CreateNode(ip, user, typ, name)
}

// CreateNode create a node
func CreateNode(ip string, user string, typ string, name string) (*Node, error) {
	key, err := sshkey()
	if err != nil {
		return nil, err
	}
	port := sshport()
	sshClient := ssh.New(ip).WithUser(user).WithKey(key).WithPort(port)

	return &Node{
		IP:   ip,
		User: user,
		Type: typ,
		Name: name,

		sshClient: sshClient,
	}, nil
}

// Provision provision node
func (n *Node) Provision(meta map[string]string) error {
	if err := n.sshClient.RunCommand(scripts["docker_version"].(string), ssh.CommandOptions{}); err != nil {

		if err := n.sshClient.RunCommand(scripts["install_docker"].(string), ssh.CommandOptions{}); err != nil {
			return err
		}

		if err := n.sshClient.RunCommand(scripts["start_dockerd"].(string), ssh.CommandOptions{}); err != nil {
			return err
		}
	}

	if n.Type == nodeTypeMaster {
		if err := n.sshClient.RunCommand(scripts["check_k3s_server"].(string), ssh.CommandOptions{}); err != nil {
			cmd := scripts["setup_k3s_master"].(func(ip string) string)(n.IP)
			if err := n.sshClient.RunCommand(cmd, ssh.CommandOptions{}); err != nil {
				return err
			}
		}
	} else if n.Type == nodeTypeAgent {
		if err := n.sshClient.RunCommand(scripts["check_k3s_agent"].(string), ssh.CommandOptions{}); err != nil {
			cmd := scripts["setup_k3s_agent"].(func(url string, tok string) string)(meta["url"], meta["token"])
			if err := n.sshClient.RunCommand(cmd, ssh.CommandOptions{}); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetToken get token from master node
func (n *Node) GetToken() (string, error) {
	if n.Type != nodeTypeMaster {
		return "", fmt.Errorf("could not get token from a non-master node")
	}
	var outPipe bytes.Buffer
	if err := n.sshClient.RunCommand(scripts["get_k3s_token"].(string), ssh.CommandOptions{Stdout: bufio.NewWriter(&outPipe)}); err != nil {
		return "", err
	}
	return outPipe.String(), nil
}

// State get node state
func (n *Node) State() {}

// Dump node information to json
func (n *Node) Dump() map[string]string {
	return map[string]string{
		"ip":   n.IP,
		"name": n.Name,
		"user": n.User,
		"type": n.Type,
	}
}

// GetType get node type
func (n *Node) GetType() string {
	return n.Type
}

// GetName get node type
func (n *Node) GetName() string {
	return n.Name
}

// GetIP get node type
func (n *Node) GetIP() string {
	return n.IP
}

// NOTE only using for unit testing
func (n *Node) setsshClient(client ssh.Clienter) {
	n.sshClient = client
}

// NOTE the reason putting sshkey() and sshport here inside node.go is because
// ssh key and ssh port is related to node it self, we may extend this in future
func sshkey() (string, error) {
	path := os.Getenv("SSH_KEY_FILE")
	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		return absPath, nil
	}

	key, err := homedir.Expand("~/.ssh/id_rsa")
	if err != nil {
		return "", err
	}
	return key, nil
}

func sshport() string {
	port := os.Getenv("SSH_PORT")
	if port != "" {
		return port
	}
	return "22"
}

var (
	_ Noder = &Node{}
)
