package client

import "net/url"

type ByocVpcEndpointService struct {
	CloudId         string `json:"cloudId"`
	Region          string `json:"region"`
	EndpointService string `json:"endpointService"`
}

func (c *Client) GetByocVpcEndpointService(cloudId, region string) (*ByocVpcEndpointService, error) {
	query := url.Values{}
	query.Set("cloudId", cloudId)
	query.Set("region", region)

	var response zillizResponse[ByocVpcEndpointService]
	if err := c.do("GET", "byoc/vpc-endpoint-service?"+query.Encode(), nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
