package cloudlayer

import (
	"encoding/json"
	"testing"
)

// TestCompositeOptions_FlatJSON verifies that composite option structs
// marshal to flat JSON without nested objects from embedding, and that
// no fields are silently dropped due to JSON tag collisions.
func TestCompositeOptions_FlatJSON(t *testing.T) {
	t.Run("URLToPDFOptions", func(t *testing.T) {
		opts := URLToPDFOptions{
			URLOptions: URLOptions{
				URL: stringPtr("https://example.com"),
			},
			PDFOptions: PDFOptions{
				PrintBackground: boolPtr(true),
			},
			PuppeteerOptions: PuppeteerOptions{
				Landscape: boolPtr(true),
			},
			BaseOptions: BaseOptions{
				Filename: stringPtr("test.pdf"),
				Async:    boolPtr(true),
			},
		}
		data, err := json.Marshal(opts)
		if err != nil {
			t.Fatal(err)
		}
		assertJSONHasKey(t, data, "url")
		assertJSONHasKey(t, data, "printBackground")
		assertJSONHasKey(t, data, "landscape")
		assertJSONHasKey(t, data, "filename")
		assertJSONHasKey(t, data, "async")
	})

	t.Run("HTMLToImageOptions", func(t *testing.T) {
		opts := HTMLToImageOptions{
			HTMLOptions: HTMLOptions{
				HTML: "PGgxPkhlbGxvPC9oMT4=",
			},
			ImageOptions: ImageOptions{
				ImageType: (*ImageType)(stringPtr("png")),
				Quality:   intPtr(90),
			},
			PuppeteerOptions: PuppeteerOptions{
				AutoScroll: boolPtr(true),
			},
			BaseOptions: BaseOptions{
				Storage: NewStorageBool(true),
			},
		}
		data, err := json.Marshal(opts)
		if err != nil {
			t.Fatal(err)
		}
		assertJSONHasKey(t, data, "html")
		assertJSONHasKey(t, data, "imageType")
		assertJSONHasKey(t, data, "quality")
		assertJSONHasKey(t, data, "autoScroll")
		assertJSONHasKey(t, data, "storage")
	})

	t.Run("TemplateToPDFOptions", func(t *testing.T) {
		opts := TemplateToPDFOptions{
			TemplateOptions: TemplateOptions{
				TemplateID: stringPtr("invoice-01"),
				Data:       map[string]interface{}{"name": "Test"},
			},
			PDFOptions: PDFOptions{
				Format: (*PDFFormat)(stringPtr("letter")),
			},
			BaseOptions: BaseOptions{
				Name: stringPtr("my-job"),
			},
		}
		data, err := json.Marshal(opts)
		if err != nil {
			t.Fatal(err)
		}
		assertJSONHasKey(t, data, "templateId")
		assertJSONHasKey(t, data, "data")
		assertJSONHasKey(t, data, "format")
		assertJSONHasKey(t, data, "name")
	})

	t.Run("MergePDFsOptions", func(t *testing.T) {
		opts := MergePDFsOptions{
			URLOptions: URLOptions{
				URL: stringPtr("https://example.com/doc.pdf"),
			},
			BaseOptions: BaseOptions{
				Async: boolPtr(true),
			},
		}
		data, err := json.Marshal(opts)
		if err != nil {
			t.Fatal(err)
		}
		assertJSONHasKey(t, data, "url")
		assertJSONHasKey(t, data, "async")
	})
}

func TestPuppeteerOptions_EmulateMediaType(t *testing.T) {
	t.Run("screen", func(t *testing.T) {
		opts := PuppeteerOptions{
			EmulateMediaType: EmulateScreen(),
		}
		data, err := json.Marshal(opts)
		if err != nil {
			t.Fatal(err)
		}
		assertJSONHasKey(t, data, "emulateMediaType")
		assertJSONKeyValue(t, data, "emulateMediaType", `"screen"`)
	})

	t.Run("null", func(t *testing.T) {
		opts := PuppeteerOptions{
			EmulateMediaType: EmulateNone(),
		}
		data, err := json.Marshal(opts)
		if err != nil {
			t.Fatal(err)
		}
		assertJSONHasKey(t, data, "emulateMediaType")
		assertJSONKeyValue(t, data, "emulateMediaType", "null")
	})

	t.Run("omitted", func(t *testing.T) {
		opts := PuppeteerOptions{
			EmulateMediaType: nil,
		}
		data, err := json.Marshal(opts)
		if err != nil {
			t.Fatal(err)
		}
		assertJSONMissingKey(t, data, "emulateMediaType")
	})
}

func TestBaseOptions_StorageOption(t *testing.T) {
	t.Run("bool true", func(t *testing.T) {
		opts := BaseOptions{Storage: NewStorageBool(true)}
		data, err := json.Marshal(opts)
		if err != nil {
			t.Fatal(err)
		}
		assertJSONKeyValue(t, data, "storage", "true")
	})

	t.Run("storage id", func(t *testing.T) {
		opts := BaseOptions{Storage: NewStorageID("my-s3")}
		data, err := json.Marshal(opts)
		if err != nil {
			t.Fatal(err)
		}
		assertJSONHasKey(t, data, "storage")
	})

	t.Run("nil", func(t *testing.T) {
		opts := BaseOptions{Storage: nil}
		data, err := json.Marshal(opts)
		if err != nil {
			t.Fatal(err)
		}
		assertJSONMissingKey(t, data, "storage")
	})
}

func TestEncodeHTML(t *testing.T) {
	html := "<h1>Hello World</h1>"
	encoded := EncodeHTML(html)
	if encoded != "PGgxPkhlbGxvIFdvcmxkPC9oMT4=" {
		t.Errorf("EncodeHTML() = %q, want PGgxPkhlbGxvIFdvcmxkPC9oMT4=", encoded)
	}
}

// --- Test helpers ---

func assertJSONHasKey(t *testing.T, data []byte, key string) {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := m[key]; !ok {
		t.Errorf("JSON missing key %q in: %s", key, data)
	}
}

func assertJSONMissingKey(t *testing.T, data []byte, key string) {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := m[key]; ok {
		t.Errorf("JSON should not have key %q, but found it in: %s", key, data)
	}
}

func assertJSONKeyValue(t *testing.T, data []byte, key, wantValue string) {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	v, ok := m[key]
	if !ok {
		t.Fatalf("JSON missing key %q in: %s", key, data)
	}
	if string(v) != wantValue {
		t.Errorf("JSON key %q = %s, want %s", key, v, wantValue)
	}
}
