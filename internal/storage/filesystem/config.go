package filesystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
)

const (
	currentConfigVersion = "v1alpha1"
	defaultConfigPath    = "flipt.json"
)

var (
	// ErrMethodNotSupported is returned when a method on the storage interface
	// is called which is not supported by the filesystem package implementation.
	// The storage implementations in this package are read-only and so each
	// mutating method should return this as the core error if invoked.
	ErrMethodNotSupported = errors.New("method not supported")
)

// Config is a structure which configures filesystem storage implementations/
// Primarily, it identifies which particular files in a target filesystem instance
// are considered and should be parsed in order to populate the storage state.
type Config struct {
	// Version identifies the current version of the configuration file
	// which also matches the resource definition version expected of
	// each identified file.
	Version    string                   `json:"version,omitempty"`
	Namespaces PathOrItemSet[Namespace] `json:"namespaces,omitempty"`
}

type Namespace struct {
	Metadata struct {
		Name string `json:"name,omitempty"`
	} `json:"metadata,omitempty"`
	Spec NamespaceSpec `json:"spec,omitempty"`
}

type NamespaceType string

const (
	StaticNamespaceType    NamespaceType = "static"
	DirectoryNamespaceType NamespaceType = "directory"
)

type NamespaceSpec struct {
	// Flags defines a set of flags within the namespace
	// Each flag may either be a static definition of the flag or a
	// path within a filesystem where the flag definition is located.
	Flags PathOrItemSet[Flag] `json:"flags,omitempty"`
}

type Flag struct {
	Metadata struct {
		Name      string `json:"name,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	} `json:"metadata,omitempty"`
	Spec FlagSpec `json:"spec,omitempty"`
}

type FlagSpec struct {
	Description string             `json:"description,omitempty"`
	Enabled     bool               `json:"enabled,omitempty"`
	Variants    map[string]Variant `json:"variants,omitempty"`
}

type Variant struct {
	Metadata struct {
		Name string `json:"name,omitempty"`
	} `json:"metadata,omitempty"`
	Spec VariantSpec `json:"spec,omitempty"`
}

type VariantSpec struct {
	Description string `json:"description,omitempty"`
	Attachment  string `json:"attachment,omitempty"`
}

type PathOrItem[T any] struct {
	Path *string
	Item *T
}

func (p *PathOrItem[T]) UnmarshalJSON(v []byte) error {
	if len(v) == 0 {
		return nil
	}

	if v[0] != '{' {
		var s string
		p.Path = &s
		return json.Unmarshal(v, p.Path)
	}

	var item T
	p.Item = &item
	return json.Unmarshal(v, p.Item)
}

type PathOrItemSet[T any] map[string]PathOrItem[T]

func (p PathOrItemSet[T]) GetItem(fs fs.FS, key string) (*T, error) {
	poi, ok := p[key]
	if !ok {
		return nil, fmt.Errorf("not found: %q", key)
	}

	if poi.Path == nil && poi.Item == nil {
		return nil, fmt.Errorf("not defined: %q", key)
	}

	if poi.Item != nil {
		return poi.Item, nil
	}

	fi, err := fs.Open(*poi.Path)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	var item T
	return &item, json.NewDecoder(fi).Decode(&item)
}
