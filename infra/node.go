package infra

import "fmt"

// Node define a node
type Node struct {
	IP   string `json:"ip"`
	User string `json:"user"`

	Type string `json:"type"`
	Name string `json:"name"`
}

// Provision provision node
func (n Node) Provision(meta map[string]string) error {
	// TODO check if docker installed or not first
	if err := InstallDocker(n); err != nil {
		return err
	}

	// TODO check if docker is running or not first
	if err := StartDockerd(n); err != nil {
		return err
	}

	// TODO check if k3s server is running or not first
	if n.Type == nodeTypeMaster {
		return ProvisionAsMaster(n)
	}

	// TODO check if k3s agent is running or not first
	return ProvisionAsAgent(meta["url"], meta["token"], n)
}

// GetToken get token from master node
func (n Node) GetToken() (string, error) {
	if n.Type != nodeTypeMaster {
		return "", fmt.Errorf("could not get token from a non-master node")
	}
	return GetToken(n)
}

// State get node state
func (n Node) State() {}

// Dump node information to json
func (n Node) Dump() {

}
