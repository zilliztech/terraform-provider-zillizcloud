package byoc_op

import "testing"

func TestIsBYOCOpProjectAgentBootstrapRequiredStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   BYOCProjectStatus
		required bool
	}{
		{name: "connected", status: BYOCProjectStatusConnected, required: false},
		{name: "creating", status: BYOCProjectStatusPending, required: false},
		{name: "running", status: BYOCProjectStatusRunning, required: false},
		{name: "init", status: BYOCProjectStatusInit, required: true},
		{name: "failed", status: BYOCProjectStatusFailed, required: true},
		{name: "deleting", status: BYOCProjectStatusDeleting, required: true},
		{name: "deleted", status: BYOCProjectStatusDeleted, required: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBYOCOpProjectAgentBootstrapRequiredStatus(int(tt.status)); got != tt.required {
				t.Fatalf("isBYOCOpProjectAgentBootstrapRequiredStatus(%d) = %v, want %v", tt.status, got, tt.required)
			}
		})
	}
}
