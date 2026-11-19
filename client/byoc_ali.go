package client

// AliParam registers customer-owned ACK resources; it contains no credentials.
type AliParam struct {
	AccountID          string   `json:"accountId"`
	ACKClusterID       string   `json:"ackClusterId"`
	VPCID              string   `json:"vpcId"`
	BucketID           string   `json:"bucketId"`
	NodeRoleARN        string   `json:"nodeRoleArn"`
	MaintenanceRoleARN string   `json:"maintenanceRoleArn"`
	StorageRoleARN     string   `json:"storageRoleArn"`
	OIDCProviderARN    string   `json:"oidcProviderArn"`
	VSwitchIDs         []string `json:"vswitchIds"`
	SecurityGroupIDs   []string `json:"securityGroupIds"`
	Zones              []string `json:"zones"`
	EndpointID         *string  `json:"endpointId,omitempty"`
}
