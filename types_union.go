package cloudlayer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LayoutDimension represents a value that can be either a string (CSS units
// like "10in", "25cm") or an integer (pixels). Use [NewLayoutDimensionString]
// or [NewLayoutDimensionInt] to create values.
type LayoutDimension struct {
	strVal *string
	intVal *int
}

// NewLayoutDimensionString creates a LayoutDimension from a CSS unit string
// (e.g., "10in", "25cm", "100px").
func NewLayoutDimensionString(s string) *LayoutDimension {
	return &LayoutDimension{strVal: &s}
}

// NewLayoutDimensionInt creates a LayoutDimension from a pixel value.
func NewLayoutDimensionInt(n int) *LayoutDimension {
	return &LayoutDimension{intVal: &n}
}

// MarshalJSON serializes the LayoutDimension as either a JSON string or number.
func (d LayoutDimension) MarshalJSON() ([]byte, error) {
	if d.strVal != nil {
		return json.Marshal(*d.strVal)
	}
	if d.intVal != nil {
		return json.Marshal(*d.intVal)
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes a JSON string or number into a LayoutDimension.
func (d *LayoutDimension) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		d.strVal = &s
		d.intVal = nil
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		d.intVal = &n
		d.strVal = nil
		return nil
	}
	return fmt.Errorf("LayoutDimension must be a string or integer, got: %s", string(data))
}

// GeneratePreviewOption represents a value that can be either a boolean or a
// [PreviewOptions] object. Use [NewGeneratePreviewBool] or
// [NewGeneratePreviewOptions] to create values.
type GeneratePreviewOption struct {
	boolVal *bool
	optsVal *PreviewOptions
}

// NewGeneratePreviewBool creates a GeneratePreviewOption from a boolean.
func NewGeneratePreviewBool(b bool) *GeneratePreviewOption {
	return &GeneratePreviewOption{boolVal: &b}
}

// NewGeneratePreviewOptions creates a GeneratePreviewOption from preview options.
func NewGeneratePreviewOptions(opts *PreviewOptions) *GeneratePreviewOption {
	return &GeneratePreviewOption{optsVal: opts}
}

// MarshalJSON serializes as either true/false or a JSON object.
func (g GeneratePreviewOption) MarshalJSON() ([]byte, error) {
	if g.boolVal != nil {
		return json.Marshal(*g.boolVal)
	}
	if g.optsVal != nil {
		return json.Marshal(g.optsVal)
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes from either a boolean or a JSON object.
func (g *GeneratePreviewOption) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		g.boolVal = &b
		g.optsVal = nil
		return nil
	}
	var opts PreviewOptions
	if err := json.Unmarshal(data, &opts); err == nil {
		g.optsVal = &opts
		g.boolVal = nil
		return nil
	}
	return fmt.Errorf("GeneratePreviewOption must be a boolean or PreviewOptions object, got: %s", string(data))
}

// StorageOption represents a value that can be either a boolean or a storage
// configuration with an ID. Use [NewStorageBool] or [NewStorageID] to create values.
type StorageOption struct {
	boolVal *bool
	idVal   *string
}

// NewStorageBool creates a StorageOption from a boolean.
// Use true to enable default storage, false to disable.
func NewStorageBool(b bool) *StorageOption {
	return &StorageOption{boolVal: &b}
}

// NewStorageID creates a StorageOption that targets a specific storage
// configuration by ID.
func NewStorageID(id string) *StorageOption {
	return &StorageOption{idVal: &id}
}

// MarshalJSON serializes as either true/false or {"id":"..."}.
func (s StorageOption) MarshalJSON() ([]byte, error) {
	if s.boolVal != nil {
		return json.Marshal(*s.boolVal)
	}
	if s.idVal != nil {
		return json.Marshal(&StorageRequestOptions{ID: *s.idVal})
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes from either a boolean or {"id":"..."}.
func (s *StorageOption) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		s.boolVal = &b
		s.idVal = nil
		return nil
	}
	var opts StorageRequestOptions
	if err := json.Unmarshal(data, &opts); err == nil {
		s.idVal = &opts.ID
		s.boolVal = nil
		return nil
	}
	return fmt.Errorf("StorageOption must be a boolean or {\"id\":\"...\"}, got: %s", string(data))
}

// NullableString represents a three-state string value:
//   - nil pointer: field is omitted from JSON (not set)
//   - non-nil with IsNull=true: field serializes as JSON null
//   - non-nil with IsNull=false: field serializes as the string value
//
// This is needed for fields like emulateMediaType where null has distinct
// meaning from omission.
type NullableString struct {
	Value  string
	IsNull bool
}

// EmulateScreen creates a NullableString with value "screen".
func EmulateScreen() *NullableString {
	return &NullableString{Value: "screen"}
}

// EmulatePrint creates a NullableString with value "print".
func EmulatePrint() *NullableString {
	return &NullableString{Value: "print"}
}

// EmulateNone creates a NullableString that serializes as JSON null,
// which disables media emulation.
func EmulateNone() *NullableString {
	return &NullableString{IsNull: true}
}

// MarshalJSON serializes the NullableString. Returns null for IsNull=true,
// or the quoted string value otherwise.
func (n NullableString) MarshalJSON() ([]byte, error) {
	if n.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}

// UnmarshalJSON deserializes from a JSON string or null.
func (n *NullableString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.IsNull = true
		n.Value = ""
		return nil
	}
	n.IsNull = false
	return json.Unmarshal(data, &n.Value)
}

// FileInput represents a file to upload via multipart form. Use
// [FileInputFromPath] for convenience, or create directly with an [io.Reader].
type FileInput struct {
	// Reader provides the file content. Required.
	Reader io.Reader
	// Filename is used as the multipart filename. Required.
	Filename string
	// ContentType is the MIME type. Defaults to "application/octet-stream" if empty.
	ContentType string
}

// FileInputFromPath opens a file at the given path and returns a FileInput.
// The caller is responsible for closing the underlying file when done
// (the returned FileInput.Reader is an *os.File).
func FileInputFromPath(path string) (*FileInput, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file %q: %w", path, err)
	}
	return &FileInput{
		Reader:   f,
		Filename: filepath.Base(path),
	}, nil
}
