package cluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSettingsDisabledRecognizesAutoscalingModes(t *testing.T) {
	dynamic := &DynamicScaling{
		Min: types.Int64Value(1),
		Max: types.Int64Value(3),
	}
	schedules := []ScheduleScaling{
		{
			Timezone: types.StringValue("UTC"),
			Cron:     types.StringValue("0 0 * * *"),
			Target:   types.Int64Value(2),
		},
	}

	tests := []struct {
		name            string
		model           ClusterResourceModel
		cuDisabled      bool
		replicaDisabled bool
	}{
		{
			name:            "nil settings",
			model:           ClusterResourceModel{},
			cuDisabled:      true,
			replicaDisabled: true,
		},
		{
			name: "dynamic settings",
			model: ClusterResourceModel{
				CuSettings:      &CuSettings{DynamicScaling: dynamic},
				ReplicaSettings: &ReplicaSettings{DynamicScaling: dynamic},
			},
			cuDisabled:      false,
			replicaDisabled: false,
		},
		{
			name: "schedule settings",
			model: ClusterResourceModel{
				CuSettings:      &CuSettings{ScheduleScaling: schedules},
				ReplicaSettings: &ReplicaSettings{ScheduleScaling: schedules},
			},
			cuDisabled:      false,
			replicaDisabled: false,
		},
		{
			name: "empty settings",
			model: ClusterResourceModel{
				CuSettings:      &CuSettings{},
				ReplicaSettings: &ReplicaSettings{},
			},
			cuDisabled:      true,
			replicaDisabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.isCuSettingsDisabled(); got != tt.cuDisabled {
				t.Fatalf("isCuSettingsDisabled() = %v, want %v", got, tt.cuDisabled)
			}
			if got := tt.model.isReplicaSettingsDisabled(); got != tt.replicaDisabled {
				t.Fatalf("isReplicaSettingsDisabled() = %v, want %v", got, tt.replicaDisabled)
			}
		})
	}
}

func TestAutoscalingRuntimeDriftDoesNotRequireFixedScaleUpdate(t *testing.T) {
	replicaSettings := &ReplicaSettings{
		DynamicScaling: &DynamicScaling{
			Min: types.Int64Value(1),
			Max: types.Int64Value(3),
		},
	}
	cuSettings := &CuSettings{
		DynamicScaling: &DynamicScaling{
			Min: types.Int64Value(2),
			Max: types.Int64Value(8),
		},
	}

	plan := ClusterResourceModel{
		CuSize:          types.Int64Value(2),
		Replica:         types.Int64Value(1),
		CuSettings:      cuSettings,
		ReplicaSettings: replicaSettings,
	}
	state := ClusterResourceModel{
		CuSize:          types.Int64Value(6),
		Replica:         types.Int64Value(3),
		CuSettings:      cuSettings,
		ReplicaSettings: replicaSettings,
	}

	if !plan.isCuSizeChanged(state) {
		t.Fatal("expected runtime cu_size drift to differ from the fixed plan value")
	}
	if !plan.isReplicaChanged(state) {
		t.Fatal("expected runtime replica drift to differ from the fixed plan value")
	}

	cuSizeChanged := plan.isCuSizeChanged(state) && plan.isCuSettingsDisabled()
	replicaChanged := plan.isReplicaChanged(state) && plan.isReplicaSettingsDisabled()
	if cuSizeChanged {
		t.Fatal("dynamic cu_settings should prevent fixed cu_size update")
	}
	if replicaChanged {
		t.Fatal("dynamic replica_settings should prevent fixed replica update")
	}
}

func TestReplicaSchemaUsesStateForUnknownWithoutDefault(t *testing.T) {
	var resp resource.SchemaResponse
	NewClusterResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	attr, ok := resp.Schema.Attributes["replica"].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("replica attribute type = %T, want schema.Int64Attribute", resp.Schema.Attributes["replica"])
	}
	if attr.Default != nil {
		t.Fatal("replica should not inject a default value")
	}
	if len(attr.PlanModifiers) == 0 {
		t.Fatal("replica should keep prior state when the planned value is unknown")
	}
}

func TestCompleteReplicaAfterCreateFillsOmittedReplica(t *testing.T) {
	state := ClusterResourceModel{
		Replica: types.Int64Unknown(),
	}
	plan := ClusterResourceModel{
		Replica: types.Int64Unknown(),
	}

	state.completeReplicaAfterCreate(plan)

	if state.Replica.IsUnknown() || state.Replica.IsNull() {
		t.Fatal("replica should be known after create")
	}
	if got := state.Replica.ValueInt64(); got != 1 {
		t.Fatalf("replica = %d, want 1", got)
	}
}

func TestCompleteReplicaAfterCreateKeepsConfiguredReplica(t *testing.T) {
	state := ClusterResourceModel{
		Replica: types.Int64Value(2),
	}
	plan := ClusterResourceModel{
		Replica: types.Int64Value(2),
	}

	state.completeReplicaAfterCreate(plan)

	if got := state.Replica.ValueInt64(); got != 2 {
		t.Fatalf("replica = %d, want 2", got)
	}
}
