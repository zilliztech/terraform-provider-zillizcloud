package client

type CloudId string

var (
	AWS          CloudId = "aws"
	Azure        CloudId = "azure"
	GCP          CloudId = "gcp"
	AliCloud     CloudId = "ali"
	TencentCloud CloudId = "tc"
)

type CloudProvider struct {
	CloudId     CloudId `json:"cloudId"`
	Description string  `json:"description"`
}

func (c *Client) ListCloudProviders() ([]CloudProvider, error) {
	var cloudProviders zillizResponse[[]CloudProvider]
	err := c.do("GET", "clouds", nil, &cloudProviders)
	return cloudProviders.Data, err
}
