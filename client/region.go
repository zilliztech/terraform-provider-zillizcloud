package client

import (
	"fmt"
	"strings"
)

type CloudRegion struct {
	ApiBaseUrl string `json:"apiBaseUrl"`
	CloudId    string `json:"cloudId"`
	RegionId   string `json:"regionId"`
}

func (c *Client) ListCloudRegions(cloudId string) ([]CloudRegion, error) {
	var cloudRegions zillizResponse[[]CloudRegion]
	err := c.do("GET", "regions?cloudId="+cloudId, nil, &cloudRegions)
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

	return fmt.Sprintf(template, cloudRegionId)
}
