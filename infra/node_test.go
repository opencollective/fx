package infra

import (
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/metrue/go-ssh-client"
	sshMocks "github.com/metrue/go-ssh-client/mocks"
	"github.com/mitchellh/go-homedir"
)

func TestGetSSHKeyFile(t *testing.T) {
	t.Run("defaut", func(t *testing.T) {
		defau, err := sshkey()
		if err != nil {
			t.Fatal(err)
		}

		defaultPath, _ := homedir.Expand("~/.ssh/id_rsa")
		if defau != defaultPath {
			t.Fatalf("should get %s but got %s", defaultPath, defau)
		}
	})

	t.Run("override from env", func(t *testing.T) {
		os.Setenv("SSH_KEY_FILE", "/tmp/id_rsa")
		keyFile, err := sshkey()
		if err != nil {
			t.Fatal(err)
		}
		if keyFile != "/tmp/id_rsa" {
			t.Fatalf("should get %s but got %s", "/tmp/id_rsa", keyFile)
		}
	})
}

func TestGetSSHPort(t *testing.T) {
	t.Run("defaut", func(t *testing.T) {
		defau := sshport()
		if defau != "22" {
			t.Fatalf("should get %s but got %s", "22", defau)
		}
	})

	t.Run("override from env", func(t *testing.T) {
		os.Setenv("SSH_PORT", "2222")
		defau := sshport()
		if defau != "2222" {
			t.Fatalf("should get %s but got %s", "2222", defau)
		}
	})
}

func TestNode(t *testing.T) {
	t.Run("master node already has docker and k3s server", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		n, err := CreateNode("127.0.0.1", "fx", nodeTypeMaster, "master")
		if err != nil {
			t.Fatal(err)
		}

		if n.sshClient == nil {
			t.Fatal("ssh client should not be nil")
		}

		sshClient := sshMocks.NewMockClienter(ctrl)
		n.setsshClient(sshClient)
		sshClient.EXPECT().RunCommand(scripts["docker_version"].(string), ssh.CommandOptions{}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["check_k3s_server"].(string), ssh.CommandOptions{}).Return(nil)
		if err := n.Provision(map[string]string{}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("master node no docker and k3s server", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		n, err := CreateNode("127.0.0.1", "fx", nodeTypeMaster, "master")
		if err != nil {
			t.Fatal(err)
		}

		if n.sshClient == nil {
			t.Fatal("ssh client should not be nil")
		}

		sshClient := sshMocks.NewMockClienter(ctrl)
		n.setsshClient(sshClient)
		sshClient.EXPECT().RunCommand(scripts["docker_version"].(string), ssh.CommandOptions{}).Return(fmt.Errorf("no such command"))
		sshClient.EXPECT().RunCommand(scripts["install_docker"].(string), ssh.CommandOptions{}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["start_dockerd"].(string), ssh.CommandOptions{}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["check_k3s_server"].(string), ssh.CommandOptions{}).Return(fmt.Errorf("no such progress"))

		cmd := scripts["setup_k3s_master"].(func(ip string) string)(n.IP)
		sshClient.EXPECT().RunCommand(cmd, ssh.CommandOptions{}).Return(nil)
		if err := n.Provision(map[string]string{}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("agent node already has docker and k3s agent", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		n, err := CreateNode("127.0.0.1", "fx", nodeTypeAgent, "agent")
		if err != nil {
			t.Fatal(err)
		}

		if n.sshClient == nil {
			t.Fatal("ssh client should not be nil")
		}

		sshClient := sshMocks.NewMockClienter(ctrl)
		n.setsshClient(sshClient)
		sshClient.EXPECT().RunCommand(scripts["docker_version"].(string), ssh.CommandOptions{}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["check_k3s_agent"].(string), ssh.CommandOptions{}).Return(nil)
		if err := n.Provision(map[string]string{}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("agent node no docker and k3s agent", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		n, err := CreateNode("127.0.0.1", "fx", nodeTypeAgent, "agent")
		if err != nil {
			t.Fatal(err)
		}

		if n.sshClient == nil {
			t.Fatal("ssh client should not be nil")
		}

		sshClient := sshMocks.NewMockClienter(ctrl)
		n.setsshClient(sshClient)
		sshClient.EXPECT().RunCommand(scripts["docker_version"].(string), ssh.CommandOptions{}).Return(fmt.Errorf("no such command"))
		sshClient.EXPECT().RunCommand(scripts["install_docker"].(string), ssh.CommandOptions{}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["start_dockerd"].(string), ssh.CommandOptions{}).Return(nil)
		sshClient.EXPECT().RunCommand(scripts["check_k3s_agent"].(string), ssh.CommandOptions{}).Return(fmt.Errorf("no such progress"))

		url := "url-1"
		token := "token-1"
		cmd := scripts["setup_k3s_agent"].(func(url string, ip string) string)(url, token)
		sshClient.EXPECT().RunCommand(cmd, ssh.CommandOptions{}).Return(nil)
		if err := n.Provision(map[string]string{"url": url, "token": token}); err != nil {
			t.Fatal(err)
		}
	})
}
