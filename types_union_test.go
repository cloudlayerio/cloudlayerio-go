package cloudlayer

import (
	"encoding/json"
	"testing"
)

func TestLayoutDimension_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		dim  *LayoutDimension
		want string
	}{
		{"string value", NewLayoutDimensionString("10in"), `"10in"`},
		{"int value", NewLayoutDimensionInt(1920), `1920`},
		{"zero int", NewLayoutDimensionInt(0), `0`},
		{"empty string", NewLayoutDimensionString(""), `""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.dim)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestLayoutDimension_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantStr *string
		wantInt *int
	}{
		{"string value", `"25cm"`, stringPtr("25cm"), nil},
		{"int value", `1080`, nil, intPtr(1080)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d LayoutDimension
			if err := json.Unmarshal([]byte(tt.input), &d); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if tt.wantStr != nil && (d.strVal == nil || *d.strVal != *tt.wantStr) {
				t.Errorf("expected string %q, got %v", *tt.wantStr, d.strVal)
			}
			if tt.wantInt != nil && (d.intVal == nil || *d.intVal != *tt.wantInt) {
				t.Errorf("expected int %d, got %v", *tt.wantInt, d.intVal)
			}
		})
	}
}

func TestLayoutDimension_RoundTrip(t *testing.T) {
	original := NewLayoutDimensionString("10in")
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded LayoutDimension
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	reencoded, err := json.Marshal(&decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(reencoded) {
		t.Errorf("round-trip mismatch: %s != %s", data, reencoded)
	}
}

func TestGeneratePreviewOption_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		opt  *GeneratePreviewOption
		want string
	}{
		{"bool true", NewGeneratePreviewBool(true), `true`},
		{"bool false", NewGeneratePreviewBool(false), `false`},
		{
			"options object",
			NewGeneratePreviewOptions(&PreviewOptions{
				Quality: 80,
				Width:   intPtr(200),
			}),
			`{"width":200,"quality":80}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.opt)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestGeneratePreviewOption_UnmarshalJSON(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		var g GeneratePreviewOption
		if err := json.Unmarshal([]byte(`true`), &g); err != nil {
			t.Fatal(err)
		}
		if g.boolVal == nil || *g.boolVal != true {
			t.Error("expected bool true")
		}
	})
	t.Run("object", func(t *testing.T) {
		var g GeneratePreviewOption
		if err := json.Unmarshal([]byte(`{"quality":90}`), &g); err != nil {
			t.Fatal(err)
		}
		if g.optsVal == nil || g.optsVal.Quality != 90 {
			t.Error("expected PreviewOptions with quality 90")
		}
	})
}

func TestStorageOption_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		opt  *StorageOption
		want string
	}{
		{"bool true", NewStorageBool(true), `true`},
		{"bool false", NewStorageBool(false), `false`},
		{"storage id", NewStorageID("abc-123"), `{"id":"abc-123"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.opt)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestStorageOption_UnmarshalJSON(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		var s StorageOption
		if err := json.Unmarshal([]byte(`false`), &s); err != nil {
			t.Fatal(err)
		}
		if s.boolVal == nil || *s.boolVal != false {
			t.Error("expected bool false")
		}
	})
	t.Run("object", func(t *testing.T) {
		var s StorageOption
		if err := json.Unmarshal([]byte(`{"id":"xyz"}`), &s); err != nil {
			t.Fatal(err)
		}
		if s.idVal == nil || *s.idVal != "xyz" {
			t.Error("expected id xyz")
		}
	})
}

func TestNullableString_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		ns   *NullableString
		want string
	}{
		{"screen", EmulateScreen(), `"screen"`},
		{"print", EmulatePrint(), `"print"`},
		{"null", EmulateNone(), `null`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.ns)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNullableString_Omitted(t *testing.T) {
	// When NullableString pointer is nil, it should be omitted from JSON
	type wrapper struct {
		Field *NullableString `json:"field,omitempty"`
	}
	w := wrapper{Field: nil}
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{}` {
		t.Errorf("nil NullableString should be omitted, got: %s", data)
	}
}

func TestNullableString_NullInJSON(t *testing.T) {
	// When NullableString is EmulateNone(), it should serialize as null in the parent
	type wrapper struct {
		Field *NullableString `json:"field,omitempty"`
	}
	w := wrapper{Field: EmulateNone()}
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"field":null}` {
		t.Errorf("EmulateNone should serialize as null, got: %s", data)
	}
}

func TestNullableString_UnmarshalJSON(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		var ns NullableString
		if err := json.Unmarshal([]byte(`"screen"`), &ns); err != nil {
			t.Fatal(err)
		}
		if ns.IsNull || ns.Value != "screen" {
			t.Error("expected value screen")
		}
	})
	t.Run("null value", func(t *testing.T) {
		var ns NullableString
		if err := json.Unmarshal([]byte(`null`), &ns); err != nil {
			t.Fatal(err)
		}
		if !ns.IsNull {
			t.Error("expected IsNull=true")
		}
	})
}
