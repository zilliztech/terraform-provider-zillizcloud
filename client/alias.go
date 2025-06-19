package client

type CreateAliasParams struct {
	DbName         string `json:"dbName"`
	AliasName      string `json:"aliasName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) CreateAlias(params *CreateAliasParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/aliases/create", params, &resp)
	if err != nil {
		return err
	}
	return nil
}

type DropAliasParams struct {
	DbName    string `json:"dbName"`
	AliasName string `json:"aliasName"`
}

func (c *ClientCollection) DropAlias(params *DropAliasParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/aliases/drop", params, &resp)
	if err != nil {
		return err
	}
	return nil
}

type ListAliasesParams struct {
	DbName         string `json:"dbName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) ListAliases(params *ListAliasesParams) ([]string, error) {
	params.DbName = c.dbName
	var resp zillizResponse[[]string]
	err := c.do("POST", "v2/vectordb/aliases/list", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

type DescribeAliasParams struct {
	DbName    string `json:"dbName"`
	AliasName string `json:"aliasName"`
}

func (c *ClientCollection) DescribeAlias(params *DescribeAliasParams) (any, error) {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/aliases/describe", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

type AlterAliasesParams struct {
	DbName         string `json:"dbName"`
	AliasName      string `json:"aliasName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) AlterAliases(params *AlterAliasesParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/aliases/alter", params, &resp)
	if err != nil {
		return err
	}
	return nil
}
