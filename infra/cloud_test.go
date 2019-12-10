package infra

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	mock_infra "github.com/metrue/fx/infra/mocks"
)

func TestLoad(t *testing.T) {
	t.Run("empty meta", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var createNodeFn = func(info map[string]string) (Noder, error) {
			return nil, nil
		}

		meta := map[string]interface{}{}
		cloud, err := Load(meta, createNodeFn)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cloud.meta, map[string]string{}) {
			t.Fatalf("should get %v but got %v", map[string]string{}, cloud.meta)
		}
	})

	t.Run("only master node", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		master := mock_infra.NewMockNoder(ctrl)
		var createNodeFn = func(info map[string]string) (Noder, error) {
			return master, nil
		}
		typ := nodeTypeMaster
		name := "master"
		ip := "127.0.0.1"
		master.EXPECT().GetName().Return(name)
		master.EXPECT().GetType().Return(typ)
		master.EXPECT().GetIP().Return(ip)
		master.EXPECT().GetConfig().Return("sample-config", nil)
		meta := map[string]interface{}{
			"nodes": map[string]map[string]string{
				"master": map[string]string{
					"type": typ,
					"ip":   ip,
					"user": "root",
					"name": "master",
				},
			},
		}
		cloud, err := Load(meta, createNodeFn)
		if err != nil {
			t.Fatal(err)
		}
		if len(cloud.nodes) != 1 {
			t.Fatalf("should get %d but got %d", 1, len(cloud.nodes))
		}

		master.EXPECT().Provision(map[string]string{}).Return(nil)
		master.EXPECT().GetToken().Return("tok-1", nil)
		if err := cloud.Provision(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("one master node and one agent", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		master := mock_infra.NewMockNoder(ctrl)
		node := mock_infra.NewMockNoder(ctrl)
		var createNodeFn = func(info map[string]string) (Noder, error) {
			if info["type"] == nodeTypeMaster {
				return master, nil
			}
			return node, nil
		}
		typ := nodeTypeMaster
		name := "master"
		ip := "127.0.0.1"
		master.EXPECT().GetName().Return(name)
		master.EXPECT().GetType().Return(typ)
		master.EXPECT().GetIP().Return(ip)
		master.EXPECT().GetConfig().Return("sample-config", nil)

		nodeType := nodeTypeAgent
		nodeName := "agent_name"
		nodeIP := "12.12.12.12"
		node.EXPECT().GetName().Return(nodeName)
		node.EXPECT().GetType().Return(nodeType).Times(2)

		meta := map[string]interface{}{
			"nodes": map[string]map[string]string{
				"master_name": map[string]string{
					"type": typ,
					"ip":   ip,
					"user": "root",
					"name": name,
				},
				"agent_name": map[string]string{
					"type": nodeType,
					"ip":   nodeIP,
					"user": "root",
					"name": nodeName,
				},
			},
		}
		cloud, err := Load(meta, createNodeFn)
		if err != nil {
			t.Fatal(err)
		}
		if len(cloud.nodes) != 2 {
			t.Fatalf("should get %d but got %d", 2, len(cloud.nodes))
		}

		master.EXPECT().Provision(map[string]string{}).Return(nil)
		master.EXPECT().GetToken().Return("tok-1", nil)
		node.EXPECT().Provision(map[string]string{
			"token":  "tok-1",
			"url":    "https://127.0.0.1:6443",
			"config": "sample-config",
		}).Return(nil)
		if err := cloud.Provision(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestProvision(t *testing.T) {}
