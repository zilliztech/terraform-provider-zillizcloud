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
	ElementDataType   string         `json:"elementDataType,omitempty"`
	IsPrimary         bool           `json:"isPrimary"`
	ElementTypeParams map[string]any `json:"elementTypeParams"`
}

type CreateCollectionParams struct {
	DbName         string           `json:"dbName"`
	CollectionName string           `json:"collectionName"`
	Schema         CollectionSchema `json:"schema"`
	Params         map[string]any   `json:"params"`
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

type CollectionDescription struct {
	Aliases            []string             `json:"aliases"`
	AutoID             bool                 `json:"autoId"`
	CollectionID       int64                `json:"collectionID"`
	CollectionName     string               `json:"collectionName"`
	ConsistencyLevel   string               `json:"consistencyLevel"`
	Description        string               `json:"description"`
	EnableDynamicField bool                 `json:"enableDynamicField"`
	Fields             []CollectionField    `json:"fields"`
	Functions          []interface{}        `json:"functions"` // Assuming functions is an array of unknown objects
	Indexes            []CollectionIndex    `json:"indexes"`
	Load               string               `json:"load"`
	PartitionsNum      int                  `json:"partitionsNum"`
	Properties         []CollectionProperty `json:"properties"`
	ShardsNum          int                  `json:"shardsNum"`
}

type CollectionField struct {
	AutoID          bool         `json:"autoId"`
	ClusteringKey   bool         `json:"clusteringKey"`
	Description     string       `json:"description"`
	ElementDataType string       `json:"elementDataType,omitempty"`
	ID              int          `json:"id"`
	Name            string       `json:"name"`
	Nullable        bool         `json:"nullable"`
	PartitionKey    bool         `json:"partitionKey"`
	PrimaryKey      bool         `json:"primaryKey"`
	Type            string       `json:"type"`
	Params          []FieldParam `json:"params"`
}

type FieldParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CollectionIndex struct {
	FieldName  string `json:"fieldName"`
	IndexName  string `json:"indexName"`
	MetricType string `json:"metricType"`
}

type CollectionProperty struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DescribeCollectionParams struct {
	DbName         string `json:"dbName"`
	CollectionName string `json:"collectionName"`
}

func (c *ClientCollection) DescribeCollection(params *DescribeCollectionParams) (*CollectionDescription, error) {
	params.DbName = c.dbName
	var resp zillizResponse[*CollectionDescription]
	err := c.do("POST", "v2/vectordb/collections/describe", params, &resp)
	if err != nil {
		return nil, err
	}
	c.logger.Debugf("DescribeCollection resp: %+v", resp.Data)
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

type AlterCollectionPropertiesParams struct {
	DbName         string         `json:"dbName"`
	CollectionName string         `json:"collectionName"`
	Properties     map[string]any `json:"properties"`
}

func (c *ClientCollection) AlterCollectionProperties(params *AlterCollectionPropertiesParams) error {
	params.DbName = c.dbName
	var resp zillizResponse[any]
	err := c.do("POST", "v2/vectordb/collections/alter_properties", params, &resp)
	if err != nil {
		return err
	}
	return nil
}
