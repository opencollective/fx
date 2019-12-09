package infra

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	sshOperator "github.com/metrue/go-ssh-client"
)

// TODO upgrade to latest when k3s fix the tls scan issue
// https://github.com/rancher/k3s/issues/556
const version = "v0.9.1"

// ProvisionAsMaster makes a master node
func ProvisionAsMaster(node Node) error {
	sshKeyFile, _ := GetSSHKeyFile()
	fmt.Println(sshKeyFile)
	ssh := sshOperator.New(node.IP).WithUser(node.User).WithKey(sshKeyFile)
	installCmd := fmt.Sprintf("curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='server --docker --tls-san %s' INSTALL_K3S_VERSION='%s' sh -", node.IP, version)
	if err := ssh.RunCommand(Sudo(installCmd, node.User), sshOperator.CommandOptions{
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}); err != nil {
		fmt.Println("setup master failed \n ===========")
		fmt.Println(err)
		fmt.Println("===========")
		return err
	}
	return nil
}

// ProvisionAsAgent makes an agent node
func ProvisionAsAgent(url string, token string, node Node) error {
	sshKeyFile, _ := GetSSHKeyFile()
	const k3sExtraArgs = "--docker"
	joinCmd := fmt.Sprintf("curl -fL https://get.k3s.io/ | K3S_URL='%s' K3S_TOKEN='%s' INSTALL_K3S_VERSION='%s' sh -s - %s", url, token, version, k3sExtraArgs)
	ssh := sshOperator.New(node.IP).WithUser(node.User).WithKey(sshKeyFile)
	if err := ssh.RunCommand(joinCmd, sshOperator.CommandOptions{
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}); err != nil {
		fmt.Println("setup agent failed \n================")
		fmt.Println(err)
		fmt.Println("================")
		return err
	}

	return nil
}

// GetToken get token from master
func GetToken(node Node) (string, error) {
	sshKeyFile, _ := GetSSHKeyFile()
	ssh := sshOperator.New(node.IP).WithUser(node.User).WithKey(sshKeyFile)
	script := "cat /var/lib/rancher/k3s/server/node-token"
	var outPipe bytes.Buffer
	if err := ssh.RunCommand(Sudo(script, node.User), sshOperator.CommandOptions{
		Stdout: bufio.NewWriter(&outPipe),
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}); err != nil {
		return "", err
	}
	return outPipe.String(), nil
}
