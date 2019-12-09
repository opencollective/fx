package infra

import (
	"fmt"
)

// Cloud define a cloud
type Cloud struct {
	meta  map[string]string
	nodes []Node
}

// Load a cloud from config
func Load(meta map[string]interface{}) (Cloud, error) {
	cloud := Cloud{
		meta:  make(map[string]string),
		nodes: []Node{},
	}
	for k, v := range meta {
		if k == "nodes" {
			nodes, ok := v.([]map[string]string)
			if !ok {
				return Cloud{}, fmt.Errorf("invalid meta")
			}
			for _, n := range nodes {
				node := Node{}
				for attr, value := range n {
					if attr == "type" {
						node.Type = value
					}
					if attr == "name" {
						node.Name = value
					}
					if attr == "ip" {
						node.IP = value
					}
					if attr == "user" {
						node.User = value
					}
				}
				cloud.nodes = append(cloud.nodes, node)
			}
		}
	}
	return cloud, nil
}

// Provision provision cloud
func (c Cloud) Provision() error {
	var master Node
	agents := []Node{}
	for _, n := range c.nodes {
		if n.Type == nodeTypeMaster {
			master = n
		} else {
			agents = append(agents, n)
		}
	}

	url := fmt.Sprintf("https://%s:6443", master.IP)
	c.meta["url"] = url
	if err := master.Provision(map[string]string{}); err != nil {
		return err
	}
	tok, err := master.GetToken()
	if err != nil {
		return err
	}
	c.meta["token"] = tok

	errCh := make(chan error, len(agents))
	defer close(errCh)

	for _, agent := range agents {
		go func(node Node) {
			errCh <- node.Provision(c.meta)
		}(agent)
	}

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// State get cloud state
func (c Cloud) State() {}

// Dump cloud information
func (c Cloud) Dump() {}
