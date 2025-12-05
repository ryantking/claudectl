package output

import (
	"encoding/json"
	"testing"
)

func TestSuccess(t *testing.T) {
	result := Success("test data")
	if !result.Success {
		t.Error("Success() should return result with Success=true")
	}
	if result.Data != "test data" {
		t.Errorf("Expected data 'test data', got %v", result.Data)
	}
	if result.Message != "" {
		t.Error("Success() should not set Message")
	}
}

func TestError(t *testing.T) {
	result := Error("test error")
	if result.Success {
		t.Error("Error() should return result with Success=false")
	}
	if result.Message != "test error" {
		t.Errorf("Expected message 'test error', got %s", result.Message)
	}
	if result.Data != nil {
		t.Error("Error() should not set Data")
	}
}

func TestResultJSON(t *testing.T) {
	result := Success(map[string]string{"key": "value"})
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var unmarshaled Result
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if !unmarshaled.Success {
		t.Error("Unmarshaled result should have Success=true")
	}
}
