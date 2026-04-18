package errors

type ResolvedEntry struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`

	FuncPath string `json:"func_path,omitempty" yaml:"func_path,omitempty"`
	FilePath string `json:"file_path,omitempty" yaml:"file_path,omitempty"`
	Line     int    `json:"line,omitempty" yaml:"line,omitempty"`

	Foreign bool `json:"foreign,omitempty" yaml:"foreign,omitempty"`
	Multi   bool `json:"multi,omitempty" yaml:"multi,omitempty"`
}

// ErrorEntry is the textual representation of an error entry in the expanded tree.
// It is used for error formatting in text, JSON and YAML.
type ErrorEntry struct {
	Resolved ResolvedEntry
	Children []*ErrorEntry
}

func newErrorEntry(message string) *ErrorEntry {
	return &ErrorEntry{
		Children: []*ErrorEntry{},
		Resolved: ResolvedEntry{
			Message: message,
		},
	}
}
