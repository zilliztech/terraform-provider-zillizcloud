package client

import (
	"testing"
)

func TestClone(t *testing.T) {
	origin, err := NewClient(WithCloudRegionId("gibberish_id"), WithApiKey("giberishkey"))
	if err != nil {
		t.Errorf("got %v, want nil", err)
	}
	client, err := origin.Clone(WithCloudRegionId("aws-west2"))
	if err != nil {
		t.Errorf("got %v, want nil", err)
	}

	expected := "https://controller.api.aws-west2.zillizcloud.com/v1/"
	if client.baseUrl != expected {
		t.Errorf("got %s, want %s", client.baseUrl, expected)
	}

}

func TestNewClient(t *testing.T) {
	type expect struct {
		baseUrl string
		err     error
	}

	testCases := []struct {
		name    string
		options []Option
		expect  expect
	}{
		{
			name: "Valid options",
			options: []Option{
				WithCloudRegionId("gibberish_id"),
				WithApiKey("gibberish_key"),
			},
			expect: expect{
				baseUrl: "https://controller.api.gibberish_id.zillizcloud.com/v1/",
				err:     nil,
			},
		},
		{
			name: "Missing API key",
			options: []Option{
				WithCloudRegionId("id"),
			},
			expect: expect{
				err: errApiKeyRequired,
			},
		},
		{
			name: "Missing cloud region ID",
			options: []Option{
				WithApiKey("key"),
			},
			expect: expect{
				baseUrl: "https://controller.api.gcp-us-west1.zillizcloud.com/v1/",
				err:     nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewClient(tc.options...)
			_ = c
			if err != tc.expect.err {
				t.Errorf("got = %v, want %v", err, tc.expect.err)
			}

			if c != nil && c.baseUrl != tc.expect.baseUrl {
				t.Errorf("got = %s, want %s", c.baseUrl, tc.expect.baseUrl)
			}
		})
	}
}
