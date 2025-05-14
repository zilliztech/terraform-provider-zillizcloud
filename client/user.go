package client

import "sort"

type ClientUser struct {
	*Client
}

func (c *Client) User(connectAddress string) (*ClientUser, error) {
	cu, err := c.cluster(connectAddress)
	if err != nil {
		return nil, err
	}
	return &ClientUser{cu}, nil
}

type Usernames []string

func (c *ClientUser) ListUsers() (Usernames, error) {
	var userResponse zillizResponse[Usernames]
	empty := map[string]any{}
	err := c.do("POST", "v2/vectordb/users/list", empty, &userResponse)
	if err != nil {
		return nil, err
	}
	sort.Strings(userResponse.Data)
	return userResponse.Data, nil
}

type CreateUserParams struct {
	Username string `json:"userName"`
	Password string `json:"password"`
}

func (c *ClientUser) CreateUser(user *CreateUserParams) error {
	var createUserResponse zillizResponse[any]
	err := c.do("POST", "v2/vectordb/users/create", user, &createUserResponse)
	if err != nil {
		return err
	}
	return nil
}

type DescribeUserParams struct {
	Username string `json:"userName"`
}

func (c *ClientUser) DescribeUser(username string) (Roles, error) {
	var userResponse zillizResponse[Roles]
	err := c.do("POST", "v2/vectordb/users/describe", &DescribeUserParams{Username: username}, &userResponse)
	if err != nil {
		return nil, err
	}
	sort.Strings(userResponse.Data)
	return userResponse.Data, nil
}

type DropUserParams struct {
	Username string `json:"userName"`
}

func (c *ClientUser) DropUser(username string) error {
	var dropUserResponse zillizResponse[any]
	return c.do("POST", "v2/vectordb/users/drop", &DropUserParams{Username: username}, &dropUserResponse)
}

type UserGrantRoleToUserParams struct {
	UserName string `json:"userName"`
	RoleName string `json:"roleName"`
}

func (c *ClientUser) GrantRoleToUser(req *UserGrantRoleToUserParams) error {
	var resp zillizResponse[any]
	return c.do("POST", "v2/vectordb/users/grant_role", req, &resp)
}

type UserRevokeRoleFromParams struct {
	UserName string `json:"userName"`
	RoleName string `json:"roleName"`
}

func (c *ClientUser) RevokeRoleFromUser(req *UserRevokeRoleFromParams) error {
	var resp zillizResponse[any]
	return c.do("POST", "v2/vectordb/users/revoke_role", req, &resp)
}
