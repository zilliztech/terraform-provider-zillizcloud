package client

import (
	"net/url"
	"strings"
)

type CloudRegion struct {
	ApiBaseUrl            string   `json:"apiBaseUrl"`
	CloudId               string   `json:"cloudId"`
	RegionId              string   `json:"regionId"`
	Domain                string   `json:"domain"`
	SupportedClusterTypes []string `json:"supportedClusterTypes"`
}

func (c *Client) ListCloudRegions(cloudId string) ([]CloudRegion, error) {
	var cloudRegions zillizResponse[[]CloudRegion]
	path := "regions"
	if cloudId != "" {
		values := url.Values{}
		values.Set("cloudId", cloudId)
		path += "?" + values.Encode()
	}
	err := c.do("GET", path, nil, &cloudRegions)
	return cloudRegions.Data, err
}

func BaseUrlFrom(cloudRegionId string) string {
	tokens := strings.Split(cloudRegionId, "-")

	cloudId := tokens[0]
	template := globalApiTemplateUrl

	switch CloudId(cloudId) {
	case AliCloud, TencentCloud:
		template = cnApiTemplateUrl
	default:
		// aws, gcp, azure, unknown will use global apiTemplateUrl

	}

	return template
}
