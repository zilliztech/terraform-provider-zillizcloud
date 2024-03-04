package client

import (
	"encoding/json"
	"strings"
	"testing"
)

var errJson = []byte(`
{
	"code":80001,
	"message":"Invalid token: Your token is not valid. Please double-check your request parameters and ensure that you have provided a valid token."
}
`)

func TestError_UnmarshalJSON(t *testing.T) {
	var e Error
	err := json.Unmarshal(errJson, &e)
	if err != nil {
		t.Errorf("Error.UnmarshalJSON() error = %v", err)
	}
	wantCode := 80001
	if e.Code != wantCode {
		t.Errorf("Error.UnmarshalJSON() = %v, want %v", e.Code, wantCode)
	}

	wantMessage := "Invalid token"
	if !strings.Contains(e.Message, wantMessage) {
		t.Errorf("Error.UnmarshalJSON() = %v, want %v", e.Message, wantMessage)
	}

}
