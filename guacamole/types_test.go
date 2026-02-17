package guacamole

import (
	"encoding/json"
	"testing"
)

func TestNullableStringMap_UnmarshalJSON(t *testing.T) {
	t.Run("null values become empty strings", func(t *testing.T) {
		var m NullableStringMap
		if err := json.Unmarshal([]byte(`{"a":null,"b":"hello","c":null}`), &m); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["a"] != "" {
			t.Errorf(`m["a"]: got %q, want ""`, m["a"])
		}
		if m["b"] != "hello" {
			t.Errorf(`m["b"]: got %q, want "hello"`, m["b"])
		}
		if m["c"] != "" {
			t.Errorf(`m["c"]: got %q, want ""`, m["c"])
		}
	})

	t.Run("empty object", func(t *testing.T) {
		var m NullableStringMap
		if err := json.Unmarshal([]byte(`{}`), &m); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("len(m): got %d, want 0", len(m))
		}
	})

	t.Run("all non-null values", func(t *testing.T) {
		var m NullableStringMap
		if err := json.Unmarshal([]byte(`{"key":"value"}`), &m); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["key"] != "value" {
			t.Errorf(`m["key"]: got %q, want "value"`, m["key"])
		}
	})
}

func TestNullableStringMap_MarshalJSON(t *testing.T) {
	// Guacamole returns HTTP 500 when the attributes field is missing or null.
	// MarshalJSON must always produce a JSON object, even for a nil map.
	t.Run("nil marshals as empty object", func(t *testing.T) {
		var m NullableStringMap // nil
		data, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != "{}" {
			t.Errorf("got %s, want {}", data)
		}
	})

	t.Run("empty map marshals as empty object", func(t *testing.T) {
		m := NullableStringMap{}
		data, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != "{}" {
			t.Errorf("got %s, want {}", data)
		}
	})

	t.Run("populated map marshals correctly", func(t *testing.T) {
		m := NullableStringMap{"hostname": "10.0.0.1", "port": "22"}
		data, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Round-trip to verify
		var got map[string]string
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal round-trip: %v", err)
		}
		if got["hostname"] != "10.0.0.1" {
			t.Errorf(`hostname: got %q, want "10.0.0.1"`, got["hostname"])
		}
		if got["port"] != "22" {
			t.Errorf(`port: got %q, want "22"`, got["port"])
		}
	})

	t.Run("nil Attributes field in struct marshals as empty object not omitted", func(t *testing.T) {
		// This is the exact scenario that caused HTTP 500 responses.
		u := User{Username: "bob", Password: "secret"}
		data, err := json.Marshal(u)
		if err != nil {
			t.Fatalf("marshal User: %v", err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("unmarshal raw: %v", err)
		}
		attrJSON, ok := raw["attributes"]
		if !ok {
			t.Fatal(`"attributes" key missing from marshaled User - Guacamole will return HTTP 500`)
		}
		if string(attrJSON) != "{}" {
			t.Errorf(`attributes: got %s, want {}`, attrJSON)
		}
	})
}
