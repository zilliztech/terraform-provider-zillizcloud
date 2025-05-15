package client

type ClientCollection struct {
	*Client
	dbName string
}

func (c *Client) Collection(connectAddress, dbName string) (*ClientCollection, error) {
	cu, err := c.cluster(connectAddress)
	if err != nil {
		return nil, err
	}
	return &ClientCollection{
		Client: cu,
		dbName: dbName,
	}, nil
}

type CollectionSchema struct {
	AutoID              bool                    `json:"autoId"`
	EnabledDynamicField bool                    `json:"enabledDynamicField"`
	Fields              []CollectionSchemaField `json:"fields"`
}
type CollectionSchemaField struct {
	FieldName         string         `json:"fieldName"`
	DataType          string         `json:"dataType"`
	IsPrimary         bool           `json:"isPrimary,omitempty"`
	ElementTypeParams map[string]any `json:"elementTypeParams,omitempty"`
}

type CreateCollectionParams struct {
	DbName         string           `json:"dbName"`
	CollectionName string           `json:"collectionName"`
	Schema         CollectionSchema `json:"schema"`
	Params         map[string]any   `json:"params,omitempty"`
}

func (c *ClientCollection) CreateCollection(params *CreateCollectionParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/collections/create", params, &resp)
	if err != nil {
		return err
	}
	return nil
}

type DropCollectionParams struct {
	DbName         string `json:"dbName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) DropCollection(params *DropCollectionParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/collections/drop", params, &resp)
	if err != nil {
		return err
	}
	return nil
}

type DescribeCollectionParams struct {
	DbName         string `json:"dbName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) DescribeCollection(params *DescribeCollectionParams) (any, error) {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/collections/describe", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

type LoadCollectionParams struct {
	DbName         string `json:"dbName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) LoadCollection(params *LoadCollectionParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/collections/load", params, &resp)
	if err != nil {
		return err
	}
	return nil
}

type ReleaseCollectionParams struct {
	DbName         string `json:"dbName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) ReleaseCollection(params *ReleaseCollectionParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/collections/release", params, &resp)
	if err != nil {
		return err
	}
	return nil
}

type ListCollectionsParams struct {
	DbName string `json:"dbName"`
}

func (c *ClientCollection) ListCollections(params *ListCollectionsParams) ([]string, error) {
	params.DbName = c.dbName
	var resp zillizResponse[[]string]
	err := c.do("POST", "v2/vectordb/collections/list", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
