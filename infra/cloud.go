package infra

import (
	"encoding/json"
	"fmt"
)

// Clouder cloud interface
type Clouder interface {
	Provision() error
	AddNode(n Noder, skipProvision bool) error
	DeleteNode(name string) error
	Dump() ([]byte, error)
}

// Cloud define a cloud
type Cloud struct {
	meta  map[string]string
	nodes map[string]Noder
}

// Load a cloud from config
func Load(meta map[string]interface{}, createNodeFns ...func(info map[string]string) (Noder, error)) (*Cloud, error) {
	cloud := &Cloud{
		meta:  make(map[string]string),
		nodes: map[string]Noder{},
	}

	createNodeFn := createNode
	if len(createNodeFns) > 0 {
		createNodeFn = createNodeFns[0]
	}

	for k, v := range meta {
		if k == "type" {
			typ, ok := v.(string)
			if ok {
				cloud.meta["type"] = typ
			}
		}
		if k == "nodes" {
			nodes, ok := v.(map[string]map[string]string)
			if !ok {
				return nil, fmt.Errorf("invalid meta")
			}
			for _, n := range nodes {
				node, err := createNodeFn(n)
				if err != nil {
					return nil, err
				}
				const skipProvision = true
				if err := cloud.AddNode(node, skipProvision); err != nil {
					return nil, err
				}
			}
		}
	}
	return cloud, nil
}

// NewCloud new a cloud
func NewCloud(typ string, node ...*Node) *Cloud {
	nodes := map[string]Noder{}
	for _, n := range node {
		nodes[n.GetName()] = n
	}
	return &Cloud{
		meta: map[string]string{
			"type": typ,
		},
		nodes: nodes,
	}
}

// Provision provision cloud
func (c *Cloud) Provision() error {
	var master Noder
	agents := []Noder{}
	for _, n := range c.nodes {
		if n.GetType() == nodeTypeMaster {
			master = n
		} else {
			agents = append(agents, n)
		}
	}

	url := fmt.Sprintf("https://%s:6443", master.GetIP())
	c.meta["url"] = url
	if err := master.Provision(map[string]string{}); err != nil {
		return err
	}

	tok, err := master.GetToken()
	if err != nil {
		return err
	}
	c.meta["token"] = tok

	if len(agents) > 0 {
		errCh := make(chan error, len(agents))
		defer close(errCh)

		for _, agent := range agents {
			go func(node Noder) {
				errCh <- node.Provision(c.meta)
			}(agent)
		}

		for range agents {
			e := <-errCh
			if e != nil {
				return err
			}
		}
	}
	return nil
}

// AddNode a node
func (c *Cloud) AddNode(n Noder, skipProvision bool) error {
	if !skipProvision {
		if err := n.Provision(c.meta); err != nil {
			return err
		}
	}

	c.nodes[n.GetName()] = n
	return nil
}

// DeleteNode a node
func (c *Cloud) DeleteNode(name string) error {
	node, ok := c.nodes[name]
	if ok {
		delete(c.nodes, name)
	}
	if node.GetType() == nodeTypeMaster && len(c.nodes) > 0 {
		return fmt.Errorf("could not delete master node since there is still agent node running")
	}
	return nil
}

// State get cloud state
func (c Cloud) State() {}

// Dump cloud information
func (c Cloud) Dump() ([]byte, error) {
	data := map[string]interface{}{
		"type": c.meta["type"],
	}

	nodes := map[string]map[string]string{}
	for name, node := range c.nodes {
		nodes[name] = node.Dump()
	}
	data["nodes"] = nodes

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return body, nil
}

var (
	_ Clouder = &Cloud{}
)
