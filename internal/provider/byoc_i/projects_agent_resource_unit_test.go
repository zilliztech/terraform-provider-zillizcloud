package byoc_op

import "testing"

func TestIsBYOCOpProjectAgentReadyStatus(t *testing.T) {
	tests := []struct {
		name   string
		status BYOCProjectStatus
		ready  bool
	}{
		{name: "connected", status: BYOCProjectStatusConnected, ready: true},
		{name: "creating", status: BYOCProjectStatusPending, ready: true},
		{name: "running", status: BYOCProjectStatusRunning, ready: true},
		{name: "init", status: BYOCProjectStatusInit, ready: false},
		{name: "failed", status: BYOCProjectStatusFailed, ready: false},
		{name: "deleting", status: BYOCProjectStatusDeleting, ready: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBYOCOpProjectAgentReadyStatus(int(tt.status)); got != tt.ready {
				t.Fatalf("isBYOCOpProjectAgentReadyStatus(%d) = %v, want %v", tt.status, got, tt.ready)
			}
		})
	}
}
