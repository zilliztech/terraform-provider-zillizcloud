package client

import (
	"context"
	"errors"
	"net/url"
	"strconv"
)

// Pointer fields preserve the exclusive access selection, including an empty project list.
type CreateServiceAccountRequest struct {
	Name               string                        `json:"name"`
	Description        string                        `json:"description"`
	OrgRoleId          *string                       `json:"orgRoleId,omitempty"`
	ProjectAssignments *[]ServiceAccountProjectRoles `json:"projectAssignments,omitempty"`
}

type CreateProjectServiceAccountRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	ProjectRoleIds []string `json:"projectRoleIds"`
}

type UpdateServiceAccountRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type ServiceAccountProjectRoles struct {
	ProjectId      string   `json:"projectId"`
	ProjectRoleIds []string `json:"projectRoleIds"`
}

type ReplaceServiceAccountAccessRequest struct {
	OrgRoleId          *string                       `json:"orgRoleId,omitempty"`
	ProjectAssignments *[]ServiceAccountProjectRoles `json:"projectAssignments,omitempty"`
}

type UpdateServiceAccountProjectRolesRequest struct {
	ProjectRoleIds []string `json:"projectRoleIds"`
}

type AssignServiceAccountRequest struct {
	ProjectRoleIds []string `json:"projectRoleIds"`
}

type ServiceAccountRole struct {
	RoleId   string `json:"roleId"`
	RoleName string `json:"roleName"`
}

type ServiceAccountProjectAssignment struct {
	ProjectId    string               `json:"projectId"`
	ProjectRoles []ServiceAccountRole `json:"projectRoles"`
}

type ServiceAccountResponse struct {
	ServiceAccountId   string                            `json:"serviceAccountId"`
	Name               string                            `json:"name"`
	Description        string                            `json:"description"`
	Status             string                            `json:"status"`
	ProjectScope       *string                           `json:"projectScope"`
	CreatedBy          string                            `json:"createdBy"`
	CreatorName        string                            `json:"creatorName"`
	CreateTime         string                            `json:"createTime"`
	OrgRoles           []ServiceAccountRole              `json:"orgRoles"`
	ProjectAssignments []ServiceAccountProjectAssignment `json:"projectAssignments"`
}

type ServiceAccountListResponse struct {
	Count           int                      `json:"count"`
	CurrentPage     int                      `json:"currentPage"`
	PageSize        int                      `json:"pageSize"`
	ServiceAccounts []ServiceAccountResponse `json:"serviceAccounts"`
}

type ServiceAccountRolesResponse struct {
	ServiceAccountId   string                            `json:"serviceAccountId"`
	OrgRoles           []ServiceAccountRole              `json:"orgRoles"`
	ProjectAssignments []ServiceAccountProjectAssignment `json:"projectAssignments"`
}

type ServiceAccountAssignmentResponse struct {
	ServiceAccountId string               `json:"serviceAccountId"`
	ProjectId        string               `json:"projectId"`
	ProjectRoles     []ServiceAccountRole `json:"projectRoles"`
}

type ServiceAccountListOptions struct {
	CurrentPage      int
	PageSize         int
	Name             string
	ServiceAccountId string
	ProjectId        string
	Status           string
}

type ServiceAccountKeyListOptions struct {
	CurrentPage int
	PageSize    int
	Name        string
	ApiKeyId    string
}

type CreateServiceAccountKeyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ExpireDays  int64  `json:"expireDays"`
}

type UpdateServiceAccountKeyRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ServiceAccountKeyResponse struct {
	ApiKeyId         string  `json:"apiKeyId"`
	ServiceAccountId string  `json:"serviceAccountId"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	ExpireTime       *string `json:"expireTime"`
	CreatedBy        string  `json:"createdBy"`
	CreatorName      string  `json:"creatorName"`
	CreateTime       string  `json:"createTime"`
}

type CreatedServiceAccountKeyResponse struct {
	ApiKey ServiceAccountKeyResponse `json:"apiKey"`
	Secret string                    `json:"secret"`
}

type ServiceAccountKeyListResponse struct {
	Count       int                         `json:"count"`
	CurrentPage int                         `json:"currentPage"`
	PageSize    int                         `json:"pageSize"`
	ApiKeys     []ServiceAccountKeyResponse `json:"apiKeys"`
}

func serviceAccountPath(id string) string {
	return "serviceAccounts/" + url.PathEscape(id)
}

func serviceAccountKeyPath(id, keyId string) string {
	return serviceAccountPath(id) + "/apiKeys/" + url.PathEscape(keyId)
}

func serviceAccountPageQuery(page, size int, name string) url.Values {
	q := url.Values{}
	if page != 0 {
		q.Set("currentPage", strconv.Itoa(page))
	}
	if size != 0 {
		q.Set("pageSize", strconv.Itoa(size))
	}
	if name != "" {
		q.Set("name", name)
	}
	return q
}

// Only explicit domain not-found errors signify disappearance. Permission errors
// must never remove an existing resource from Terraform state.
func IsServiceAccountNotFound(err error) bool {
	return serviceAccountErrorCode(err, 21155)
}

func IsServiceAccountKeyNotFound(err error) bool {
	return serviceAccountErrorCode(err, 21159) || IsServiceAccountNotFound(err)
}

func serviceAccountErrorCode(err error, code int) bool {
	var pointer *Error
	if errors.As(err, &pointer) {
		return pointer.Code == code
	}
	var value Error
	return errors.As(err, &value) && value.Code == code
}

func (c *Client) CreateServiceAccount(ctx context.Context, req *CreateServiceAccountRequest) (*ServiceAccountResponse, error) {
	var response zillizResponse[ServiceAccountResponse]
	if err := c.doContext(ctx, "POST", "serviceAccounts", req, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) CreateProjectServiceAccount(ctx context.Context, projectId string, req *CreateProjectServiceAccountRequest) (*ServiceAccountResponse, error) {
	var response zillizResponse[ServiceAccountResponse]
	if err := c.doContext(ctx, "POST", "project/"+url.PathEscape(projectId)+"/serviceAccounts", req, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) GetServiceAccount(ctx context.Context, id string) (*ServiceAccountResponse, error) {
	var response zillizResponse[ServiceAccountResponse]
	if err := c.doWithRetry(ctx, "GET", serviceAccountPath(id), nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) UpdateServiceAccount(ctx context.Context, id string, req *UpdateServiceAccountRequest) (*ServiceAccountResponse, error) {
	var response zillizResponse[ServiceAccountResponse]
	// A combined metadata + disable request cannot be replayed: once disabled,
	// the server rejects further metadata edits. The Terraform resource sends
	// metadata and status separately, so its updates use the retryable path.
	do := c.doWithRetry
	if req != nil && req.Status != nil && (req.Name != nil || req.Description != nil) {
		do = c.doContext
	}
	if err := do(ctx, "PATCH", serviceAccountPath(id), req, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) AssignServiceAccount(ctx context.Context, id, projectId string, req *AssignServiceAccountRequest) (*ServiceAccountAssignmentResponse, error) {
	var response zillizResponse[ServiceAccountAssignmentResponse]
	if err := c.doWithRetry(ctx, "POST", serviceAccountPath(id)+"/projectAssignments/"+url.PathEscape(projectId), req, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) ReplaceServiceAccountAccess(ctx context.Context, id string, req *ReplaceServiceAccountAccessRequest) (*ServiceAccountRolesResponse, error) {
	var response zillizResponse[ServiceAccountRolesResponse]
	if err := c.doWithRetry(ctx, "PUT", serviceAccountPath(id)+"/access", req, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) UpdateServiceAccountProjectRoles(ctx context.Context, id, projectId string, req *UpdateServiceAccountProjectRolesRequest) (*ServiceAccountAssignmentResponse, error) {
	var response zillizResponse[ServiceAccountAssignmentResponse]
	if err := c.doWithRetry(ctx, "PATCH", serviceAccountPath(id)+"/projectAssignments/"+url.PathEscape(projectId), req, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) CreateServiceAccountKey(ctx context.Context, id string, req *CreateServiceAccountKeyRequest) (*CreatedServiceAccountKeyResponse, error) {
	var response zillizResponse[CreatedServiceAccountKeyResponse]
	if err := c.doContext(ctx, "POST", serviceAccountPath(id)+"/apiKeys", req, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) GetServiceAccountKey(ctx context.Context, id, keyId string) (*ServiceAccountKeyResponse, error) {
	var response zillizResponse[ServiceAccountKeyResponse]
	if err := c.doWithRetry(ctx, "GET", serviceAccountKeyPath(id, keyId), nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) UpdateServiceAccountKey(ctx context.Context, id, keyId string, req *UpdateServiceAccountKeyRequest) (*ServiceAccountKeyResponse, error) {
	var response zillizResponse[ServiceAccountKeyResponse]
	if err := c.doWithRetry(ctx, "PATCH", serviceAccountKeyPath(id, keyId), req, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) DeleteServiceAccount(ctx context.Context, id string) error {
	var response zillizResponse[bool]
	return c.doWithRetry(ctx, "DELETE", serviceAccountPath(id), nil, &response)
}

func (c *Client) UnassignServiceAccount(ctx context.Context, id, projectId string) error {
	var response zillizResponse[bool]
	return c.doWithRetry(ctx, "DELETE", serviceAccountPath(id)+"/projectAssignments/"+url.PathEscape(projectId), nil, &response)
}

func (c *Client) DeleteServiceAccountKey(ctx context.Context, id, keyId string) error {
	var response zillizResponse[bool]
	return c.doWithRetry(ctx, "DELETE", serviceAccountKeyPath(id, keyId), nil, &response)
}

func (c *Client) ListServiceAccounts(ctx context.Context, options ServiceAccountListOptions) (*ServiceAccountListResponse, error) {
	q := serviceAccountPageQuery(options.CurrentPage, options.PageSize, options.Name)
	if options.ServiceAccountId != "" {
		q.Set("serviceAccountId", options.ServiceAccountId)
	}
	if options.ProjectId != "" {
		q.Set("projectId", options.ProjectId)
	}
	if options.Status != "" {
		q.Set("status", options.Status)
	}
	var response zillizResponse[ServiceAccountListResponse]
	if err := c.doWithRetry(ctx, "GET", "serviceAccounts?"+q.Encode(), nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) ListServiceAccountKeys(ctx context.Context, id string, options ServiceAccountKeyListOptions) (*ServiceAccountKeyListResponse, error) {
	q := serviceAccountPageQuery(options.CurrentPage, options.PageSize, options.Name)
	if options.ApiKeyId != "" {
		q.Set("apiKeyId", options.ApiKeyId)
	}
	var response zillizResponse[ServiceAccountKeyListResponse]
	if err := c.doWithRetry(ctx, "GET", serviceAccountPath(id)+"/apiKeys?"+q.Encode(), nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
