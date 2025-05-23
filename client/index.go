package client

type CreateIndexParams struct {
	DbName         string        `json:"dbName"`
	CollectionName string        `json:"collectionName"`
	IndexParams    []IndexParams `json:"indexParams"`
}
type IndexParams struct {
	MetricType  string            `json:"metricType"`
	FieldName   string            `json:"fieldName"`
	IndexName   string            `json:"indexName"`
	IndexConfig map[string]string `json:"indexConfig"`
}

func (c *ClientCollection) CreateIndex(params *CreateIndexParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/indexes/create", params, &resp)
	if err != nil {
		return err
	}
	return nil
}

type DropIndexParams struct {
	DbName         string `json:"dbName"`
	CollectionName string `json:"collectionName"`
	IndexName      string `json:"indexName"`
}

func (c *ClientCollection) DropIndex(params *DropIndexParams) error {
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/indexes/drop", params, &resp)
	if err != nil {
		return err
	}
	return nil
}

type DescribeIndexParams struct {
	DbName         string `json:"dbName"`
	CollectionName string `json:"collectionName"`
	IndexName      string `json:"indexName"`
}

func (c *ClientCollection) DescribeIndex(params *DescribeIndexParams) (any, error) {
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/indexes/describe", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

type ListIndexParams struct {
	DbName         string `json:"dbName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) ListIndex(params *ListIndexParams) ([]string, error) {
	var resp zillizResponse[[]string]
	err := c.do("POST", "v2/vectordb/indexes/list", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
