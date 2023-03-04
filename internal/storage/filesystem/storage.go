package filesystem

import (
	"context"
	"io/fs"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
)

type Store struct {
	*FlagStore
}

func NewStore(fs fs.FS) *Store {
	return &Store{
		FlagStore: NewFlagStore(fs),
	}
}

func (s *Store) GetRule(ctx context.Context, id string) (*flipt.Rule, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) ListRules(ctx context.Context, flagKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Rule], error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) CountRules(ctx context.Context) (uint64, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	panic("not implemented") // TODO: Implement
}

func (s *Store) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	panic("not implemented") // TODO: Implement
}

func (s *Store) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	panic("not implemented") // TODO: Implement
}

func (s *Store) GetSegment(ctx context.Context, key string) (*flipt.Segment, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) ListSegments(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Segment], error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) CountSegments(ctx context.Context) (uint64, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	panic("not implemented") // TODO: Implement
}

func (s *Store) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	panic("not implemented") // TODO: Implement
}

// GetEvaluationRules returns rules applicable to flagKey provided
// Note: Rules MUST be returned in order by Rank
func (s *Store) GetEvaluationRules(ctx context.Context, flagKey string) ([]*storage.EvaluationRule, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) String() string {
	return "filesystem"
}
