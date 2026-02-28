package client

import "fmt"

// BackupPolicyParams represents the parameters for creating or updating a backup policy
type BackupPolicyParams struct {
	Frequency         string            `json:"frequency"`
	StartTime         string            `json:"startTime"`
	RetentionDays     int               `json:"retentionDays"`
	Enabled           bool              `json:"enabled"`
	CrossRegionCopies []CrossRegionCopy `json:"crossRegionCopies"`
}

// CrossRegionCopy represents a cross-region backup copy configuration
type CrossRegionCopy struct {
	RegionId      string `json:"regionId"`
	RetentionDays int    `json:"retentionDays"`
}

// BackupPolicy represents the backup policy response from the API
type BackupPolicy struct {
	Frequency         string            `json:"frequency"`
	StartTime         string            `json:"startTime"`
	RetentionDays     int               `json:"retentionDays"`
	Status            string            `json:"status"`
	CrossRegionCopies []CrossRegionCopy `json:"crossRegionCopies"`
}

// UpsertBackupPolicy creates or updates a backup policy for a cluster
func (c *Client) UpsertBackupPolicy(clusterId string, params *BackupPolicyParams) error {
	var response zillizResponse[any]

	if params == nil {
		return fmt.Errorf("params is nil")
	}
	if params.CrossRegionCopies == nil {
		params.CrossRegionCopies = []CrossRegionCopy{}
	}
	err := c.do("POST", "clusters/"+clusterId+"/backups/policy", params, &response)
	if err != nil {
		return err
	}
	return nil
}

// GetBackupPolicy retrieves the backup policy for a cluster
func (c *Client) GetBackupPolicy(clusterId string) (*BackupPolicy, error) {
	var response zillizResponse[BackupPolicy]
	err := c.do("GET", "clusters/"+clusterId+"/backups/policy", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

// DeleteBackupPolicy deletes the backup policy for a cluster by disabling it
func (c *Client) DeleteBackupPolicy(clusterId string) error {
	params := &BackupPolicyParams{
		Enabled: false,
	}
	var response zillizResponse[any]
	err := c.do("POST", "clusters/"+clusterId+"/backups/policy", params, &response)
	if err != nil {
		return err
	}
	return nil
}
