package client

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitByocOpTunnelClientImageWireField(t *testing.T) {
	for _, tc := range []struct {
		name     string
		opConfig string
		want     string
	}{
		{name: "configured", opConfig: `{"tunnelClientImageUrl":"cloud-tunnel-client:test"}`, want: "cloud-tunnel-client:test"},
		{name: "missing", opConfig: `{}`},
		{name: "null", opConfig: `{"tunnelClientImageUrl":null}`},
		{name: "empty", opConfig: `{"tunnelClientImageUrl":""}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var settings GetByocOpProjectSettingsResponse
			require.NoError(t, json.Unmarshal([]byte(`{"opConfig":`+tc.opConfig+`}`), &settings))
			require.Equal(t, tc.want, settings.OpConfig.TunnelClientImageUrl)
			var project DescribeByocOpProjectResponse
			require.NoError(t, json.Unmarshal([]byte(`{"vpcOpConfig":`+tc.opConfig+`}`), &project))
			require.NotNil(t, project.OpConfig)
			require.Equal(t, tc.want, project.OpConfig.TunnelClientImageUrl)
		})
	}
}
