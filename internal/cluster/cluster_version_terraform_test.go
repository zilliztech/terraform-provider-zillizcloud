package cluster_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func writeVersionTestResponse(t *testing.T, w http.ResponseWriter, status int, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data}); err != nil {
		t.Errorf("write response: %v", err)
	}
}

func versionTestProvider(server *httptest.Server) string {
	return fmt.Sprintf(`
provider "zillizcloud" {
  api_key      = "test-key"
  host_address = %q
  region_id    = "aws-us-east-2"
}
`, server.URL+"/v2")
}

func TestTerraformClusterVersionsDataSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/clusters/versions" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			http.NotFound(w, r)
			return
		}
		query := r.URL.Query()
		switch {
		case query.Get("regionId") == "aws-us-east-2" && !query.Has("backupId"):
			writeVersionTestResponse(t, w, http.StatusOK, json.RawMessage(`[{"regionId":"aws-us-east-2","versions":[{"milvusVersion":"3.0.x","channel":"PUBLIC_PREVIEW"}]}]`))
		case query.Get("backupId") == "backup-1" && !query.Has("regionId"):
			writeVersionTestResponse(t, w, http.StatusOK, json.RawMessage(`[{"regionId":"gcp-us-west1","versions":[{"milvusVersion":"2.6.x","channel":"ORIGIN"}]}]`))
		case query.Get("backupId") == "backup-1" && query.Get("regionId") == "aws-us-east-2":
			writeVersionTestResponse(t, w, http.StatusOK, []any{})
		default:
			t.Errorf("unexpected query: %s", query.Encode())
			http.Error(w, "unexpected query", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: versionTestProvider(server) + `
data "zillizcloud_cluster_versions" "create" {
  region_id = "aws-us-east-2"
}
data "zillizcloud_cluster_versions" "restore" {
  backup_id = "backup-1"
}
data "zillizcloud_cluster_versions" "mismatch" {
  region_id = "aws-us-east-2"
  backup_id = "backup-1"
}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.zillizcloud_cluster_versions.create", "regions.0.versions.0.milvus_version", "3.0.x"),
				resource.TestCheckResourceAttr("data.zillizcloud_cluster_versions.create", "regions.0.versions.0.channel", "PUBLIC_PREVIEW"),
				resource.TestCheckResourceAttr("data.zillizcloud_cluster_versions.restore", "regions.0.region_id", "gcp-us-west1"),
				resource.TestCheckResourceAttr("data.zillizcloud_cluster_versions.restore", "regions.0.versions.0.channel", "ORIGIN"),
				resource.TestCheckResourceAttr("data.zillizcloud_cluster_versions.mismatch", "regions.#", "0"),
			),
		}},
	})
}

func TestTerraformDedicatedClusterVersionSelection(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit=%t", explicit), func(t *testing.T) {
			var creates atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method + " " + r.URL.Path {
				case "GET /v2/clusters/versions":
					writeVersionTestResponse(t, w, http.StatusOK, json.RawMessage(`[{"regionId":"aws-us-east-2","versions":[{"milvusVersion":"3.0.x","channel":"PUBLIC_PREVIEW"}]}]`))
				case "POST /v2/clusters", "POST /v2/clusters/createDedicated":
					creates.Add(1)
					var body map[string]json.RawMessage
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("decode create: %v", err)
					}
					_, hasVersion := body["version"]
					if explicit {
						if r.URL.Path != "/v2/clusters" || string(body["version"]) != `{"milvusVersion":"3.0.x","channel":"PUBLIC_PREVIEW"}` {
							t.Errorf("explicit selection sent to %s with version %s", r.URL.Path, body["version"])
						}
						if string(body["bucketInfo"]) != `{"bucketName":"test-bucket","prefix":"test-prefix"}` || string(body["keyIdentifier"]) != `"arn:aws:kms:us-east-2:123456789012:key/test"` {
							t.Errorf("versioned create did not preserve the BYOC bucket and key")
						}
					} else if r.URL.Path != "/v2/clusters/createDedicated" || hasVersion {
						t.Errorf("legacy create sent to %s (version present: %t)", r.URL.Path, hasVersion)
					}
					if string(body["regionId"]) != `"aws-us-east-2"` {
						t.Errorf("wrong region: %s", body["regionId"])
					}
					writeVersionTestResponse(t, w, http.StatusAccepted, json.RawMessage(`{"clusterId":"in01-target","username":"db_admin","password":"test","prompt":"created"}`))
				case "GET /v2/clusters/in01-target":
					writeVersionTestResponse(t, w, http.StatusOK, json.RawMessage(`{"clusterId":"in01-target","clusterName":"version-test","projectId":"project-1","regionId":"aws-us-east-2","plan":"Enterprise","cuType":"Performance-optimized","cuSize":1,"replica":1,"status":"RUNNING","milvusVersion":"3.0.x"}`))
				case "GET /v2/clusters/in01-target/labels":
					writeVersionTestResponse(t, w, http.StatusOK, json.RawMessage(`{"labels":{}}`))
				case "DELETE /v2/clusters/in01-target/drop":
					writeVersionTestResponse(t, w, http.StatusOK, json.RawMessage(`{"clusterId":"in01-target"}`))
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			config := versionTestProvider(server)
			clusterConfig := `
resource "zillizcloud_cluster" "test" {
  cluster_name   = "version-test"
  project_id     = "project-1"
  region_id      = "aws-us-east-2"
  plan           = "Enterprise"
  cu_size        = 1
  desired_status = "RUNNING"
  %s
}
data "zillizcloud_cluster" "test" {
  id = zillizcloud_cluster.test.id
}
`
			selection := `version = { milvus_version = "3.0.x", channel = "PUBLIC_PREVIEW" }`
			input := ""
			if explicit {
				config += `
data "zillizcloud_cluster_versions" "available" {
  region_id = "aws-us-east-2"
}
`
				input = `
  version         = data.zillizcloud_cluster_versions.available.regions[0].versions[0]
  bucket_info     = { bucket_name = "test-bucket", prefix = "test-prefix" }
  aws_cse_key_arn  = "arn:aws:kms:us-east-2:123456789012:key/test"
`
			}
			steps := []resource.TestStep{{
				Config: config + fmt.Sprintf(clusterConfig, input),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "milvus_version", "3.0.x"),
					resource.TestCheckResourceAttr("data.zillizcloud_cluster.test", "milvus_version", "3.0.x"),
				),
				PreventPostDestroyRefresh: true,
			}}
			if explicit {
				selection = `version = { milvus_version = "2.6.x", channel = "GA" }`
			} else {
				steps = append(steps, resource.TestStep{
					ResourceName:            "zillizcloud_cluster.test",
					ImportState:             true,
					ImportStateId:           "in01-target,aws-us-east-2",
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"password", "prompt", "username", "timeouts"},
				})
			}
			replacement := config + fmt.Sprintf(clusterConfig, selection)
			replacementPlan := regexp.MustCompile(`(?s)zillizcloud_cluster.test must be replaced.*Plan: 1 to add, 0 to change, 1 to destroy`)
			steps = append(steps, resource.TestStep{
				Config: replacement, PlanOnly: true, ExpectError: replacementPlan,
			}, resource.TestStep{
				Config:   strings.Replace(replacement, "project-1", "project-2", 1),
				PlanOnly: true, ExpectError: replacementPlan,
			})
			resource.UnitTest(t, resource.TestCase{ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories, Steps: steps})
			if creates.Load() != 1 {
				t.Errorf("create requests = %d, want 1", creates.Load())
			}
		})
	}
}

const restoreAddress = "zillizcloud_cluster_restore.test"
const restoreRequestPath = "/v2/clusters/in01-source/backups/backup-1/restoreCluster"
const restoreTargetPath = "/v2/clusters/in01-target"

func restoreTestConfig(server *httptest.Server, custom string) string {
	return versionTestProvider(server) + fmt.Sprintf(`
resource "zillizcloud_cluster_restore" "test" {
  source_cluster_id = "in01-source"
  backup_id         = "backup-1"
  project_id        = "project-1"
  cluster_name      = "restored"
  cu_size           = 1
  %s
}
`, custom)
}

func writeMissingRestoreTarget(t *testing.T, w http.ResponseWriter) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(map[string]any{"code": 40021}); err != nil {
		t.Error(err)
	}
}

func TestTerraformRestoreTimeoutKeepsAcceptedTarget(t *testing.T) {
	var creates atomic.Int32
	var completed atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "POST " + restoreRequestPath:
			creates.Add(1)
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if string(body["version"]) != `{"milvusVersion":"2.6.x","channel":"ORIGIN"}` || body["restoreVersionPolicy"] != nil {
				t.Errorf("expected an explicit ORIGIN restore")
			}
			writeVersionTestResponse(t, w, http.StatusOK, json.RawMessage(`{"clusterId":"in01-target","jobId":"job-1"}`))
		case "GET /v2/jobs/job-1":
			status := "IN_PROGRESS"
			if completed.Load() {
				status = "SUCCESSFUL"
			}
			writeVersionTestResponse(t, w, http.StatusOK, map[string]string{"status": status})
		case "GET " + restoreTargetPath:
			if !completed.Load() {
				writeMissingRestoreTarget(t, w)
				return
			}
			writeVersionTestResponse(t, w, http.StatusOK, json.RawMessage(`{"clusterId":"in01-target","projectId":"project-1","clusterName":"restored","cuSize":1,"regionId":"aws-us-east-2","status":"RUNNING"}`))
		case "DELETE " + restoreTargetPath + "/drop":
			writeVersionTestResponse(t, w, http.StatusOK, nil)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := restoreTestConfig(server, `
  version = { milvus_version = "2.6.x", channel = "ORIGIN" }
  timeouts { create = "1s" }
`)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: strings.Replace(config, "timeouts {", `restore_version_policy = "ORIGINAL"
  timeouts {`, 1),
				PlanOnly: true, ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(restoreAddress, "id", "in01-target"),
					resource.TestCheckResourceAttr(restoreAddress, "job_id", "job-1"),
					resource.TestCheckResourceAttr(restoreAddress, "restore_status", "IN_PROGRESS"),
				),
			},
			{
				PreConfig: func() { completed.Store(true) },
				Config:    config,
				Check:     resource.TestCheckResourceAttr(restoreAddress, "restore_status", "SUCCESSFUL"),
			},
		},
	})
	if creates.Load() != 1 {
		t.Errorf("restore requests = %d, want 1", creates.Load())
	}
}

func TestTerraformTerminalRestoreFailure(t *testing.T) {
	for _, status := range []string{"FAILED", "CANCELED"} {
		t.Run(status, func(t *testing.T) {
			var creates, deletes atomic.Int32
			var targetMissing atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method + " " + r.URL.Path {
				case "POST " + restoreRequestPath:
					creates.Add(1)
					targetMissing.Store(false)
					var body map[string]json.RawMessage
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if string(body["restoreVersionPolicy"]) != `"ORIGINAL"` || body["version"] != nil {
						t.Error("expected policy-only restore")
					}
					writeVersionTestResponse(t, w, http.StatusOK, json.RawMessage(`{"clusterId":"in01-target","jobId":"job-1"}`))
				case "GET /v2/jobs/job-1":
					jobStatus := status
					if creates.Load() > 1 {
						jobStatus = "SUCCESSFUL"
					}
					writeVersionTestResponse(t, w, http.StatusOK, map[string]string{"status": jobStatus})
				case "GET " + restoreTargetPath:
					if targetMissing.Load() {
						writeMissingRestoreTarget(t, w)
						return
					}
					writeVersionTestResponse(t, w, http.StatusOK, json.RawMessage(`{"clusterId":"in01-target","projectId":"project-1","clusterName":"restored","cuSize":1,"regionId":"aws-us-east-2","status":"RUNNING"}`))
				case "DELETE " + restoreTargetPath + "/drop":
					deletes.Add(1)
					writeVersionTestResponse(t, w, http.StatusOK, nil)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			config := restoreTestConfig(server, `restore_version_policy = "ORIGINAL"`)
			failure := regexp.MustCompile("job-1 for cluster in01-target reached " + status)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: config, ExpectError: failure},
					{Config: config, ExpectError: failure},
					{
						PreConfig: func() {
							if creates.Load() != 1 || deletes.Load() != 0 {
								t.Errorf("failed target changed unexpectedly: %d creates, %d deletes", creates.Load(), deletes.Load())
							}
							targetMissing.Store(true)
						},
						Config: config,
						Check:  resource.TestCheckResourceAttr(restoreAddress, "restore_status", "SUCCESSFUL"),
					},
				},
			})
			if creates.Load() != 2 || deletes.Load() != 1 {
				t.Errorf("expected one retry and one teardown delete, got %d creates and %d deletes", creates.Load(), deletes.Load())
			}
		})
	}
}

func TestTerraformRestoredClusterLifecycle(t *testing.T) {
	var mu sync.Mutex
	var creates, modifies, resizes, deletes int
	name, description, cuSize := "restored", "initial", 1
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.Method + " " + r.URL.Path {
		case "POST " + restoreRequestPath:
			creates++
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["collectionStatus"] != "RELEASE" {
				t.Errorf("wrong collection status: %v", body["collectionStatus"])
			}
			writeVersionTestResponse(t, w, http.StatusOK, map[string]string{"clusterId": "in01-target", "jobId": "job-1"})
		case "GET /v2/jobs/job-1":
			writeVersionTestResponse(t, w, http.StatusOK, map[string]string{"status": "SUCCESSFUL"})
		case "GET " + restoreTargetPath:
			writeVersionTestResponse(t, w, http.StatusOK, map[string]any{
				"clusterId": "in01-target", "projectId": "project-1", "regionId": "aws-us-east-2",
				"milvusVersion": "2.6.x", "status": "RUNNING", "connectAddress": "https://cluster.example",
				"clusterName": name, "description": description, "cuSize": cuSize,
			})
		case "POST " + restoreTargetPath + "/modifyProperties":
			modifies++
			var body struct {
				Name        string  `json:"clusterName"`
				Description *string `json:"description"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body.Name != "" {
				name = body.Name
			}
			if body.Description != nil {
				description = *body.Description
			}
			writeVersionTestResponse(t, w, http.StatusOK, nil)
		case "POST " + restoreTargetPath + "/modify":
			resizes++
			var body struct {
				CUSize int `json:"cuSize"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			cuSize = body.CUSize
			writeVersionTestResponse(t, w, http.StatusOK, nil)
		case "DELETE " + restoreTargetPath + "/drop":
			deletes++
			writeVersionTestResponse(t, w, http.StatusOK, nil)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := restoreTestConfig(server, `
  description = "initial"
  version = { milvus_version = "2.6.x", channel = "ORIGIN" }
  timeouts { create = "45m" }
`) + `
resource "terraform_data" "dependent" {
  triggers_replace = [zillizcloud_cluster_restore.test.id, zillizcloud_cluster_restore.test.job_id]
}
`
	explicitDefault := strings.Replace(config, "timeouts {", `collection_status = "RELEASE"
  timeouts {`, 1)
	updatedTimeout := strings.Replace(explicitDefault, "45m", "60m", 1)
	updatedCluster := strings.NewReplacer(
		`cluster_name      = "restored"`, `cluster_name      = "renamed"`,
		"cu_size           = 1", "cu_size           = 2", `description = "initial"`, `description = ""`,
	).Replace(updatedTimeout)
	planFor := func(action plancheck.ResourceActionType) resource.ConfigPlanChecks {
		return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(restoreAddress, action),
			plancheck.ExpectResourceAction("terraform_data.dependent", plancheck.ResourceActionNoop),
		}}
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: resource.TestCheckResourceAttr(restoreAddress, "collection_status", "RELEASE")},
			{Config: explicitDefault, ConfigPlanChecks: planFor(plancheck.ResourceActionNoop)},
			{Config: updatedTimeout, ConfigPlanChecks: planFor(plancheck.ResourceActionUpdate)},
			{
				Config: updatedCluster, ConfigPlanChecks: planFor(plancheck.ResourceActionUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(restoreAddress, "id", "in01-target"),
					resource.TestCheckResourceAttr(restoreAddress, "cluster_name", "renamed"),
					resource.TestCheckResourceAttr(restoreAddress, "description", ""),
					resource.TestCheckResourceAttr(restoreAddress, "cu_size", "2"),
				),
			},
			{
				Config: updatedCluster,
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					name, description, cuSize = "console-name", "console-description", 3
				},
				ConfigPlanChecks: planFor(plancheck.ResourceActionUpdate),
			},
			{
				ResourceName: restoreAddress, ImportState: true, ImportStateVerify: true,
				ImportStateId:           "in01-target,in01-source,backup-1,job-1,RELEASE,2.6.x:ORIGIN",
				ImportStateVerifyIgnore: []string{"username", "password", "timeouts"},
			},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	if creates != 1 || modifies != 2 || resizes != 2 || deletes != 1 {
		t.Errorf("expected one restore, two in-place modifications and resizes, and one teardown delete, got %d, %d, %d, %d",
			creates, modifies, resizes, deletes)
	}
}
