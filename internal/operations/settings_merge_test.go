package operations

import (
	"reflect"
	"testing"
)

func TestMergeSettingsSmart(t *testing.T) {
	base := map[string]interface{}{
		"key1": "value1",
		"key2": map[string]interface{}{
			"nested1": "nested_value1",
			"nested2": "nested_value2",
		},
		"key3": []interface{}{"item1", "item2"},
	}

	overlay := map[string]interface{}{
		"key1": "overlay_value1",
		"key2": map[string]interface{}{
			"nested2": "overlay_nested_value2",
			"nested3": "nested_value3",
		},
		"key3": []interface{}{"item2", "item3"},
		"key4": "new_value",
	}

	merged := MergeSettingsSmart(base, overlay)

	// Check scalar override
	if merged["key1"] != "overlay_value1" {
		t.Errorf("Expected key1 to be 'overlay_value1', got %v", merged["key1"])
	}

	// Check nested merge
	nested, ok := merged["key2"].(map[string]interface{})
	if !ok {
		t.Fatal("key2 should be a map")
	}
	if nested["nested1"] != "nested_value1" {
		t.Error("nested1 should be preserved")
	}
	if nested["nested2"] != "overlay_nested_value2" {
		t.Error("nested2 should be overridden")
	}
	if nested["nested3"] != "nested_value3" {
		t.Error("nested3 should be added")
	}

	// Check array merge (should deduplicate)
	arr, ok := merged["key3"].([]interface{})
	if !ok {
		t.Fatal("key3 should be a slice")
	}
	expectedItems := []interface{}{"item1", "item2", "item3"}
	if !reflect.DeepEqual(arr, expectedItems) {
		t.Errorf("Expected array %v, got %v", expectedItems, arr)
	}

	// Check new key
	if merged["key4"] != "new_value" {
		t.Error("key4 should be added")
	}
}

func TestMergeLists(t *testing.T) {
	base := []interface{}{"a", "b", "c"}
	overlay := []interface{}{"b", "c", "d"}

	merged := mergeLists(base, overlay)
	expected := []interface{}{"a", "b", "c", "d"}

	if !reflect.DeepEqual(merged, expected) {
		t.Errorf("Expected %v, got %v", expected, merged)
	}
}

func TestLoadSaveJSONSettings(t *testing.T) {
	original := map[string]interface{}{
		"key1": "value1",
		"key2": map[string]interface{}{
			"nested": "value",
		},
		"key3": []interface{}{float64(1), float64(2), float64(3)}, // JSON numbers become float64
	}

	data, err := SaveJSONSettings(original)
	if err != nil {
		t.Fatalf("SaveJSONSettings failed: %v", err)
	}

	loaded, err := LoadJSONSettings(data)
	if err != nil {
		t.Fatalf("LoadJSONSettings failed: %v", err)
	}

	// JSON unmarshals numbers as float64, so we need to compare carefully
	if loaded["key1"] != original["key1"] {
		t.Error("key1 doesn't match")
	}
	if !reflect.DeepEqual(loaded["key2"], original["key2"]) {
		t.Error("key2 doesn't match")
	}
	// For arrays, JSON numbers become float64
	loadedArr := loaded["key3"].([]interface{})
	originalArr := original["key3"].([]interface{})
	if len(loadedArr) != len(originalArr) {
		t.Error("Array length doesn't match")
	}
}
