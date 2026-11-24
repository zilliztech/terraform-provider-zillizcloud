package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceAccountEndpoints(t *testing.T) {
	empty := []string{}
	orgRole := "r-billing-admin"
	assignments := []ServiceAccountProjectRoles{}
	description := ""
	cases := []struct {
		name, method, path, body, response string
		call                               func(*Client) error
	}{
		{
			"create org",
			"POST",
			"/v2/serviceAccounts",
			`{"name":"test","description":"","orgRoleId":"r-billing-admin"}`,
			`{"serviceAccountId":"sa-test"}`,
			func(c *Client) error {
				v, e := c.CreateServiceAccount(context.Background(), &CreateServiceAccountRequest{
					Name:      "test",
					OrgRoleId: &orgRole,
				})
				if e == nil {
					require.Equal(t, "sa-test", v.ServiceAccountId)
				}
				return e
			},
		},
		{
			"create empty assignments",
			"POST",
			"/v2/serviceAccounts",
			`{"name":"test","description":"","projectAssignments":[]}`,
			`{}`,
			func(c *Client) error {
				_, e := c.CreateServiceAccount(context.Background(), &CreateServiceAccountRequest{
					Name:               "test",
					ProjectAssignments: &assignments,
				})
				return e
			},
		},
		{
			"create local",
			"POST",
			"/v2/project/proj-a/serviceAccounts",
			`{"name":"local","description":"","projectRoleIds":[]}`,
			`{}`,
			func(c *Client) error {
				_, e := c.CreateProjectServiceAccount(context.Background(), "proj-a", &CreateProjectServiceAccountRequest{
					Name:           "local",
					ProjectRoleIds: empty,
				})
				return e
			},
		},
		{
			"list",
			"GET",
			"/v2/serviceAccounts?currentPage=2&name=a%26b+%2B&projectId=proj-a&pageSize=100&serviceAccountId=sa&status=active",
			"",
			`{"count":2,"currentPage":2,"pageSize":100,"serviceAccounts":[]}`,
			func(c *Client) error {
				v, e := c.ListServiceAccounts(context.Background(), ServiceAccountListOptions{
					CurrentPage:      2,
					PageSize:         100,
					Name:             "a&b +",
					ProjectId:        "proj-a",
					ServiceAccountId: "sa",
					Status:           "active",
				})
				if e == nil {
					require.Equal(t, 2, v.Count)
				}
				return e
			},
		},
		{
			"get",
			"GET",
			"/v2/serviceAccounts/sa-test",
			"",
			`{"projectScope":null}`,
			func(c *Client) error {
				v, e := c.GetServiceAccount(context.Background(), "sa-test")
				if e == nil {
					require.Nil(t, v.ProjectScope)
				}
				return e
			},
		},
		{
			"edit",
			"PATCH",
			"/v2/serviceAccounts/sa-test",
			`{"description":""}`,
			`{}`,
			func(c *Client) error {
				_, e := c.UpdateServiceAccount(context.Background(), "sa-test", &UpdateServiceAccountRequest{Description: &description})
				return e
			},
		},
		{
			"delete",
			"DELETE",
			"/v2/serviceAccounts/sa-test",
			"",
			`true`,
			func(c *Client) error { return c.DeleteServiceAccount(context.Background(), "sa-test") },
		},
		{
			"assign",
			"POST",
			"/v2/serviceAccounts/sa-test/projectAssignments/proj-a",
			`{"projectRoleIds":[]}`,
			`{}`,
			func(c *Client) error {
				_, e := c.AssignServiceAccount(context.Background(), "sa-test", "proj-a", &AssignServiceAccountRequest{ProjectRoleIds: empty})
				return e
			},
		},
		{
			"unassign",
			"DELETE",
			"/v2/serviceAccounts/sa-test/projectAssignments/proj-a",
			"",
			`true`,
			func(c *Client) error { return c.UnassignServiceAccount(context.Background(), "sa-test", "proj-a") },
		},
		{
			"replace with org role", "PUT", "/v2/serviceAccounts/sa-test/access",
			`{"orgRoleId":"r-billing-admin"}`, `{"orgRoles":[{"roleId":"r-billing-admin"}]}`,
			func(c *Client) error {
				v, e := c.ReplaceServiceAccountAccess(context.Background(), "sa-test", &ReplaceServiceAccountAccessRequest{OrgRoleId: &orgRole})
				if e == nil {
					require.Equal(t, "r-billing-admin", v.OrgRoles[0].RoleId)
				}
				return e
			},
		},
		{
			"replace with project access", "PUT", "/v2/serviceAccounts/sa-test/access",
			`{"projectAssignments":[{"projectId":"proj-a","projectRoleIds":[]}]}`, `{}`,
			func(c *Client) error {
				projects := []ServiceAccountProjectRoles{{ProjectId: "proj-a", ProjectRoleIds: empty}}
				_, e := c.ReplaceServiceAccountAccess(context.Background(), "sa-test", &ReplaceServiceAccountAccessRequest{ProjectAssignments: &projects})
				return e
			},
		},
		{
			"clear all access", "PUT", "/v2/serviceAccounts/sa-test/access",
			`{"projectAssignments":[]}`, `{}`,
			func(c *Client) error {
				_, e := c.ReplaceServiceAccountAccess(context.Background(), "sa-test", &ReplaceServiceAccountAccessRequest{ProjectAssignments: &assignments})
				return e
			},
		},
		{
			"clear project roles", "PATCH", "/v2/serviceAccounts/sa-test/projectAssignments/proj-a",
			`{"projectRoleIds":[]}`, `{"serviceAccountId":"sa-test","projectId":"proj-a","projectRoles":[]}`,
			func(c *Client) error {
				v, e := c.UpdateServiceAccountProjectRoles(context.Background(), "sa-test", "proj-a", &UpdateServiceAccountProjectRolesRequest{ProjectRoleIds: empty})
				if e == nil {
					require.Equal(t, "proj-a", v.ProjectId)
					require.Empty(t, v.ProjectRoles)
				}
				return e
			},
		},
		{
			"create key",
			"POST",
			"/v2/serviceAccounts/sa-test/apiKeys",
			`{"name":"key","description":"","expireDays":90}`,
			`{"apiKey":{"apiKeyId":"key-test"},"secret":"once"}`,
			func(c *Client) error {
				v, e := c.CreateServiceAccountKey(context.Background(), "sa-test", &CreateServiceAccountKeyRequest{
					Name:       "key",
					ExpireDays: 90,
				})
				if e == nil {
					require.Equal(t, "once", v.Secret)
					require.Equal(t, "key-test", v.ApiKey.ApiKeyId)
				}
				return e
			},
		},
		{
			"list keys",
			"GET",
			"/v2/serviceAccounts/sa-test/apiKeys?apiKeyId=key&currentPage=2&name=a%26b&pageSize=100",
			"",
			`{"count":3,"apiKeys":[]}`,
			func(c *Client) error {
				v, e := c.ListServiceAccountKeys(context.Background(), "sa-test", ServiceAccountKeyListOptions{
					CurrentPage: 2,
					PageSize:    100,
					Name:        "a&b",
					ApiKeyId:    "key",
				})
				if e == nil {
					require.Equal(t, 3, v.Count)
				}
				return e
			},
		},
		{
			"get key",
			"GET",
			"/v2/serviceAccounts/sa-test/apiKeys/key-test",
			"",
			`{"expireTime":null}`,
			func(c *Client) error {
				v, e := c.GetServiceAccountKey(context.Background(), "sa-test", "key-test")
				if e == nil {
					require.Nil(t, v.ExpireTime)
				}
				return e
			},
		},
		{
			"edit key",
			"PATCH",
			"/v2/serviceAccounts/sa-test/apiKeys/key-test",
			`{"description":""}`,
			`{}`,
			func(c *Client) error {
				_, e := c.UpdateServiceAccountKey(context.Background(), "sa-test", "key-test", &UpdateServiceAccountKeyRequest{Description: &description})
				return e
			},
		},
		{
			"delete key",
			"DELETE",
			"/v2/serviceAccounts/sa-test/apiKeys/key-test",
			"",
			`true`,
			func(c *Client) error { return c.DeleteServiceAccountKey(context.Background(), "sa-test", "key-test") },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tc.method, r.Method)
				// Parse expected query separately so parameter ordering is immaterial.
				expected, err := http.NewRequest(tc.method, "http://example"+tc.path, nil)
				require.NoError(t, err)
				require.Equal(t, expected.URL.Path, r.URL.Path)
				require.Equal(t, expected.URL.Query(), r.URL.Query())
				require.Equal(t, "Bearer test", r.Header.Get("Authorization"))
				if tc.body != "" {
					var body json.RawMessage
					require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					require.JSONEq(t, tc.body, string(body))
				}
				fmt.Fprintf(w, `{"code":0,"data":%s}`, tc.response)
			}))
			defer server.Close()
			c, err := NewClient(WithApiKey("test"), WithBaseUrl(server.URL+"/v2"))
			require.NoError(t, err)
			require.NoError(t, tc.call(c))
		})
	}
}

func TestServiceAccountNotFound(t *testing.T) {
	for _, httpStatus := range []int{
		200,
		404,
		403,
	} {
		for _, code := range []int{
			21155,
			21159,
			403,
			404,
			21156,
		} {
			t.Run(fmt.Sprintf("%d-%d", httpStatus, code), func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(httpStatus)
					fmt.Fprintf(w, `{"code":%d,"message":"test"}`, code)
				}))
				defer server.Close()
				c, err := NewClient(WithApiKey("test"), WithBaseUrl(server.URL))
				require.NoError(t, err)
				_, err = c.GetServiceAccount(context.Background(), "sa-test")
				require.Error(t, err)
				require.Equal(t, code == 21155, IsServiceAccountNotFound(err))
				require.Equal(t, code == 21155 || code == 21159, IsServiceAccountKeyNotFound(err))
			})
		}
	}
	require.False(t, IsServiceAccountNotFound(nil))
	require.False(t, IsServiceAccountNotFound(fmt.Errorf("connection failed")))
}
