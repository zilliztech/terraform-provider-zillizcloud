package client

type ClientCluster struct {
	*Client
}

func (c *Client) Cluster(connectAddress string) (*ClientCluster, error) {
	cu, err := c.cluster(connectAddress)
	if err != nil {
		return nil, err
	}
	return &ClientCluster{cu}, nil
}

type CreateDatabaseParams struct {
	DbName     string         `json:"dbName"`
	Properties map[string]any `json:"properties"`
}

func (c *ClientCluster) CreateDatabase(params CreateDatabaseParams) (any, error) {
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/databases/create", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

type DescribeDatabaseParams struct {
	DbName string `json:"dbName"`
}

type DescribeDatabaseResponse struct {
	DbName     string           `json:"dbName"`
	Properties []map[string]any `json:"properties"`
}

func (c *ClientCluster) DescribeDatabase(params DescribeDatabaseParams) (*DescribeDatabaseResponse, error) {
	var resp zillizResponse[*DescribeDatabaseResponse]
	err := c.do("POST", "v2/vectordb/databases/describe", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

type DropDatabaseParams struct {
	DbName string `json:"dbName"`
}

func (c *ClientCluster) DropDatabase(params DropDatabaseParams) (any, error) {
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/databases/drop", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

type UpdateDatabaseParams struct {
	DbName     string         `json:"dbName"`
	Properties map[string]any `json:"properties"`
}

func (c *ClientCluster) UpdateDatabase(params UpdateDatabaseParams) (any, error) {
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/databases/alter", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *ClientCluster) ListDatabases() ([]string, error) {
	var resp zillizResponse[[]string]
	err := c.do("POST", "v2/vectordb/databases/list", map[string]any{}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
