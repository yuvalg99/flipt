package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strconv"

	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FlagStore struct {
	fs fs.FS

	configPath string
}

func WithConfigPath(p string) containers.Option[FlagStore] {
	return func(s *FlagStore) {
		s.configPath = p
	}
}

func NewFlagStore(fs fs.FS, opts ...containers.Option[FlagStore]) *FlagStore {
	store := &FlagStore{
		fs:         fs,
		configPath: defaultConfigPath,
	}

	containers.ApplyAll(store, opts...)

	return store
}

func (f *FlagStore) config() (c Config, _ error) {
	fi, err := f.fs.Open(f.configPath)
	if err != nil {
		return c, err
	}

	return c, json.NewDecoder(fi).Decode(&c)
}

func (f *FlagStore) GetFlag(ctx context.Context, key string) (*flipt.Flag, error) {
	config, err := f.config()
	if err != nil {
		return nil, err
	}

	ns, err := config.Namespaces.GetItem(f.fs, "default")
	if err != nil {
		return nil, err
	}

	flag, err := ns.Spec.Flags.GetItem(f.fs, key)
	if err != nil {
		return nil, err
	}

	return flagToRPCFlag(key, flag), nil
}

func (f *FlagStore) ListFlags(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error) {
	var (
		params storage.QueryParams
		res    storage.ResultSet[*flipt.Flag]
	)

	config, err := f.config()
	if err != nil {
		return res, err
	}

	ns, err := config.Namespaces.GetItem(f.fs, "default")
	if err != nil {
		return res, err
	}

	for _, opt := range opts {
		opt(&params)
	}

	params.Normalize()

	// parse all flags as they stored in an random order map
	// then sort them by name based in the defined order.
	for key, item := range ns.Spec.Flags {
		res.Results = append(res.Results, flagToRPCFlag(key, item.Item))
	}

	fn := func(i, j int) bool {
		return res.Results[i].Key > res.Results[j].Key
	}

	if params.Order != storage.OrderAsc {
		less := fn
		fn = func(i, j int) bool { return !less(i, j) }
	}

	sort.Slice(res.Results, fn)

	// paginate the result based on query params.
	offset := params.Offset
	if params.PageToken != "" {
		offset, err = strconv.ParseUint(params.PageToken, 10, 64)
		if err != nil {
			return storage.ResultSet[*flipt.Flag]{}, err
		}
	}

	if offset > 0 {
		if int(offset) >= len(res.Results) {
			return storage.ResultSet[*flipt.Flag]{}, nil
		}

		res.Results = res.Results[offset:]
	}

	if params.Limit > 0 {
		if int(params.Limit) >= len(res.Results) {
			// we must be on the last page so return all
			// results and leave next page token blank
			return res, nil
		}

		// otherwise, we limit results and set a next page token
		res.Results = res.Results[:params.Limit]
		res.NextPageToken = fmt.Sprintf("%d", offset+params.Limit)
	}

	return res, nil
}

func (f *FlagStore) CountFlags(ctx context.Context) (uint64, error) {
	config, err := f.config()
	if err != nil {
		return 0, err
	}

	ns, err := config.Namespaces.GetItem(f.fs, "default")
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.Spec.Flags)), nil
}

func flagToRPCFlag(key string, flag *Flag) *flipt.Flag {
	fflag := &flipt.Flag{
		Key:         key,
		Name:        flag.Metadata.Name,
		Enabled:     flag.Spec.Enabled,
		Description: flag.Spec.Description,
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
	}

	for key, variant := range flag.Spec.Variants {
		fflag.Variants = append(fflag.Variants, &flipt.Variant{
			Id:          key,
			FlagKey:     fflag.Key,
			Key:         key,
			Name:        variant.Metadata.Name,
			Description: variant.Spec.Description,
			Attachment:  variant.Spec.Attachment,
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
		})
	}

	return fflag
}

// Note: the following ensure FlagStore implements storage.FlagStore
// The fileystem implementations of storage are read-only.
// This is why the following metgods all return a not supported error.

func (f *FlagStore) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	return nil, fmt.Errorf("flags: CreateFlag: %w", ErrMethodNotSupported)
}

func (f *FlagStore) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	return nil, fmt.Errorf("flags: UpdateFlag: %w", ErrMethodNotSupported)
}

func (f *FlagStore) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	return fmt.Errorf("flags: DeleteFlag: %w", ErrMethodNotSupported)
}

func (f *FlagStore) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	return nil, fmt.Errorf("flags: CreateVariant: %w", ErrMethodNotSupported)
}

func (f *FlagStore) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	return nil, fmt.Errorf("flags: UpdateVariant: %w", ErrMethodNotSupported)
}

func (f *FlagStore) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	return fmt.Errorf("flags: DeleteVariant: %w", ErrMethodNotSupported)
}
