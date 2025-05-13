package client

import "sort"

type ClientRole struct {
	*Client
}

func (c *Client) Role(connectAddress string) (*ClientRole, error) {
	cu, err := c.Cluster(connectAddress)
	if err != nil {
		return nil, err
	}
	return &ClientRole{cu}, nil
}

type Roles []string

func (c *ClientRole) ListRoles() (Roles, error) {
	var rolesResponse zillizResponse[Roles]
	empty := map[string]any{}
	err := c.do("POST", "v2/vectordb/roles/list", empty, &rolesResponse)
	if err != nil {
		return nil, err
	}
	sort.Strings(rolesResponse.Data)
	return rolesResponse.Data, nil
}
