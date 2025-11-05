package client

import (
	"testing"
)

func TestClient_ListCloudProviders(t *testing.T) {

	t.Run("ListCloudProviders", func(t *testing.T) {

		checkFn := make(map[string]func(cloudProviders []CloudProvider) bool, 0)

		has := func(cloudId string) func(cloudProviders []CloudProvider) bool {
			return func(cloudProviders []CloudProvider) bool {
				for _, cp := range cloudProviders {
					if string(cp.CloudId) == cloudId {
						return true
					}
				}
				return false
			}
		}

		checkFn["aws"] = has("aws")
		checkFn["azure"] = has("azure")
		checkFn["gcp"] = has("gcp")

		c, teardown := zillizClient[[]CloudProvider](t)
		defer teardown()

		got, err := c.ListCloudProviders()
		if err != nil {
			t.Fatalf("failed to ListCloudProviders: %v", err)
		}

		for k, fn := range checkFn {
			if !fn(got) {
				t.Errorf("has %s failed", k)
			}
		}

	})
}

func TestClient_UpgradeProjectPlan(t *testing.T) {

	t.Run("UpgradeProjectPlan", func(t *testing.T) {
		c, teardown := zillizClient[string](t)
		defer teardown()

		projectId := "proj-77e7a5474cff4fcf457fbe"
		plan := "Enterprise"
		_, err := c.UpgradeProjectPlan(projectId, plan)
		if err != nil {
			t.Fatalf("failed to upgrade project plan: %v", err)
		}
		t.Logf("upgraded project plan to %s", plan)
	})
}

func TestClient_UpgradeProjectPlan_failed(t *testing.T) {

	t.Run("UpgradeProjectPlan_failed", func(t *testing.T) {
		c, teardown := zillizClient[string](t)
		defer teardown()

		projectId := "proj-77e7a5474cff4fcf457fbe"
		plan := "Enterprise"
		_, err := c.UpgradeProjectPlan(projectId, plan)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		t.Logf("failed to upgrade project plan to %s", plan)
	})
}
