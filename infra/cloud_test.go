package infra

import (
	"reflect"
	"testing"
)

func TestCloud(t *testing.T) {
	t.Run("empty meta", func(t *testing.T) {
		meta := map[string]interface{}{}
		cloud, err := Load(meta)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cloud.meta, map[string]string{}) {
			t.Fatalf("should get %v but got %v", map[string]string{}, cloud.meta)
		}
	})
	t.Run("only master node", func(t *testing.T) {
		meta := map[string]interface{}{
			"nodes": []map[string]string{
				map[string]string{
					"type": "master",
					"ip":   "43.224.35.195",
					"user": "root",
					"name": "master",
				},
			},
		}
		cloud, err := Load(meta)
		if err != nil {
			t.Fatal(err)
		}
		if len(cloud.nodes) != 1 {
			t.Fatalf("should get %d but got %d", 1, len(cloud.nodes))
		}

		if err := cloud.Provision(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("one master node and one agent", func(t *testing.T) {
		meta := map[string]interface{}{
			"nodes": []map[string]string{
				map[string]string{
					"type": "master",
					"ip":   "43.224.35.195",
					"user": "root",
					"name": "master",
				},
				map[string]string{
					"type": "agent",
					"ip":   "43.224.35.195",
					"user": "root",
					"name": "master",
				},
			},
		}
		cloud, err := Load(meta)
		if err != nil {
			t.Fatal(err)
		}
		if len(cloud.nodes) != 2 {
			t.Fatalf("should get %d but got %d", 2, len(cloud.nodes))
		}

		if err := cloud.Provision(); err != nil {
			t.Fatal(err)
		}
	})
}
