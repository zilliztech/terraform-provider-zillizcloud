package client

type CreatePartitionsParams struct {
	DbName         string `json:"dbName"`
	PartitionsName string `json:"partitionName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) CreatePartitions(params *CreatePartitionsParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/partitions/create", params, &resp)
	if err != nil {
		return err
	}
	return nil
}

type DropPartitionsParams struct {
	DbName         string `json:"dbName"`
	PartitionsName string `json:"partitionName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) DropPartitions(params *DropPartitionsParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/partitions/drop", params, &resp)
	if err != nil {
		return err
	}
	return nil
}

type ListPartitionsesParams struct {
	DbName         string `json:"dbName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) ListPartitionses(params *ListPartitionsesParams) ([]string, error) {
	params.DbName = c.dbName
	var resp zillizResponse[[]string]
	err := c.do("POST", "v2/vectordb/partitions/list", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

type DescribePartitionsParams struct {
	DbName         string `json:"dbName"`
	PartitionsName string `json:"partitionName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) DescribePartitions(params *DescribePartitionsParams) (any, error) {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/partitions/describe", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
