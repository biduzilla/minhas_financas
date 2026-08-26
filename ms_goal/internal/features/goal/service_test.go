package goal

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"ms_goal/internal/core/domain/apiError"
	"ms_goal/internal/core/filters"
	"ms_goal/internal/core/messaging/events"
)

type mockGoalRepo struct {
	findAllFn  func(ctx context.Context, search string, status GoalStatus, f filters.Filters) ([]*Goal, filters.Metadata, error)
	findByIdFn func(ctx context.Context, id uuid.UUID) (*Goal, error)
	insertFn   func(ctx context.Context, model *Goal) error
	updateFn   func(ctx context.Context, model *Goal) error
	deleteFn   func(ctx context.Context, id uuid.UUID) error
}

func (m *mockGoalRepo) FindAll(ctx context.Context, search string, status GoalStatus, f filters.Filters) ([]*Goal, filters.Metadata, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, search, status, f)
	}
	return nil, filters.Metadata{}, nil
}

func (m *mockGoalRepo) FindByID(ctx context.Context, id uuid.UUID) (*Goal, error) {
	if m.findByIdFn != nil {
		return m.findByIdFn(ctx, id)
	}
	return nil, nil
}

func (m *mockGoalRepo) Insert(ctx context.Context, model *Goal) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, model)
	}
	return nil
}

func (m *mockGoalRepo) Update(ctx context.Context, model *Goal) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, model)
	}
	return nil
}

func (m *mockGoalRepo) DeleteById(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockCache struct {
	getFn            func(ctx context.Context, key string, dest any) error
	setFn            func(ctx context.Context, key string, value any, ttl *time.Duration) error
	deleteFn         func(ctx context.Context, keys ...string) error
	deleteByPrefixFn func(ctx context.Context, prefix string) error
}

func (m *mockCache) Get(ctx context.Context, key string, dest any) error {
	if m.getFn != nil {
		return m.getFn(ctx, key, dest)
	}
	return errors.New("cache miss")
}

func (m *mockCache) Set(ctx context.Context, key string, value any, ttl *time.Duration) error {
	if m.setFn != nil {
		return m.setFn(ctx, key, value, ttl)
	}
	return nil
}

func (m *mockCache) Delete(ctx context.Context, keys ...string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, keys...)
	}
	return nil
}

func (m *mockCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	if m.deleteByPrefixFn != nil {
		return m.deleteByPrefixFn(ctx, prefix)
	}
	return nil
}

type mockKeyBuilder struct{}

func (m *mockKeyBuilder) BuildItemKey(id string) string     { return "goal:" + id }
func (m *mockKeyBuilder) BuildListKey(params ...any) string { return "goal:list" }
func (m *mockKeyBuilder) GetPrefix() string                 { return "goal:" }

type mockWriteExecutor struct {
	executeFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockWriteExecutor) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.executeFn != nil {
		return m.executeFn(ctx, fn)
	}
	return fn(ctx) // comportamento padrão: apenas executa
}

type mockGoalProducer struct {
	publishCreatedErr error
	publishDeletedErr error
	createdEvent      *events.GoalEvent
	deletedEvent      *events.GoalEvent
}

func (m *mockGoalProducer) PublishGoalCreated(ctx context.Context, event events.GoalEvent) error {
	m.createdEvent = &event
	return m.publishCreatedErr
}

func (m *mockGoalProducer) PublishGoalDeleted(ctx context.Context, event events.GoalEvent) error {
	m.deletedEvent = &event
	return m.publishDeletedErr
}

func newValidGoal() *Goal {
	deadline := time.Now().AddDate(0, 1, 0)
	return &Goal{
		ID:            uuid.New(),
		UserID:        uuid.New(),
		Name:          "Comprar carro",
		TargetAmount:  100000,
		CurrentAmount: 0,
		Status:        GoalStatusInProgress,
		Deadline:      deadline,
	}
}

func newTestService(
	repo *mockGoalRepo,
	cacheMock *mockCache,
	kb *mockKeyBuilder,
	we *mockWriteExecutor,
	producer *mockGoalProducer,
) *GoalService {
	return NewService(repo, cacheMock, kb, we, producer)
}

func TestFindByID_Success_FromDB(t *testing.T) {
	goal := newValidGoal()
	repo := &mockGoalRepo{
		findByIdFn: func(ctx context.Context, id uuid.UUID) (*Goal, error) {
			return goal, nil
		},
	}
	cacheMock := &mockCache{
		getFn: func(ctx context.Context, key string, dest any) error {
			return errors.New("cache miss")
		},
	}
	kb := &mockKeyBuilder{}
	we := &mockWriteExecutor{}
	producer := &mockGoalProducer{}

	svc := newTestService(repo, cacheMock, kb, we, producer)

	result, err := svc.FindByID(context.Background(), goal.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.ID != goal.ID {
		t.Errorf("Expected ID %s, got %s", goal.ID, result.ID)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	repo := &mockGoalRepo{
		findByIdFn: func(ctx context.Context, id uuid.UUID) (*Goal, error) {
			return nil, apiError.ErrRecordNotFound
		},
	}
	cacheMock := &mockCache{
		getFn: func(ctx context.Context, key string, dest any) error {
			return errors.New("cache miss")
		},
	}
	kb := &mockKeyBuilder{}
	we := &mockWriteExecutor{}
	producer := &mockGoalProducer{}

	svc := newTestService(repo, cacheMock, kb, we, producer)

	_, err := svc.FindByID(context.Background(), uuid.New())

	if !errors.Is(err, apiError.ErrRecordNotFound) {
		t.Errorf("Expected ErrRecordNotFound, got %v", err)
	}
}

func TestFindAll_Success_FromDB(t *testing.T) {
	goals := []*Goal{newValidGoal()}
	expectedMeta := filters.Metadata{TotalRecords: 1, CurrentPage: 1, PageSize: 10}
	repo := &mockGoalRepo{
		findAllFn: func(ctx context.Context, search string, status GoalStatus, f filters.Filters) ([]*Goal, filters.Metadata, error) {
			return goals, expectedMeta, nil
		},
	}
	cacheMock := &mockCache{
		getFn: func(ctx context.Context, key string, dest any) error {
			return errors.New("cache miss")
		},
	}
	kb := &mockKeyBuilder{}
	we := &mockWriteExecutor{}
	producer := &mockGoalProducer{}

	svc := newTestService(repo, cacheMock, kb, we, producer)

	resultGoals, meta, err := svc.FindAll(context.Background(), "", GoalStatusInProgress, filters.Filters{Page: 1, PageSize: 10, Sort: "id"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(resultGoals) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(resultGoals))
	}
	if meta.TotalRecords != 1 {
		t.Errorf("Expected TotalRecords 1, got %d", meta.TotalRecords)
	}
}

func TestFindAll_Error(t *testing.T) {
	repo := &mockGoalRepo{
		findAllFn: func(ctx context.Context, search string, status GoalStatus, f filters.Filters) ([]*Goal, filters.Metadata, error) {
			return nil, filters.Metadata{}, errors.New("query failed")
		},
	}
	cacheMock := &mockCache{
		getFn: func(ctx context.Context, key string, dest any) error {
			return errors.New("cache miss")
		},
	}
	kb := &mockKeyBuilder{}
	we := &mockWriteExecutor{}
	producer := &mockGoalProducer{}

	svc := newTestService(repo, cacheMock, kb, we, producer)

	_, _, err := svc.FindAll(context.Background(), "", GoalStatusInProgress, filters.Filters{Page: 1, PageSize: 10, Sort: "id"})

	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestInsert_ValidationFailure(t *testing.T) {
	invalidGoal := newValidGoal()
	invalidGoal.Name = ""

	insertCalled := false
	repo := &mockGoalRepo{
		insertFn: func(ctx context.Context, model *Goal) error {
			insertCalled = true
			return nil
		},
	}
	cacheMock := &mockCache{}
	kb := &mockKeyBuilder{}
	we := &mockWriteExecutor{}
	producer := &mockGoalProducer{}

	svc := newTestService(repo, cacheMock, kb, we, producer)

	err := svc.Insert(context.Background(), invalidGoal)

	if err == nil {
		t.Fatal("Expected validation error")
	}
	if insertCalled {
		t.Error("Repo Insert should NOT be called for invalid goal")
	}
}

func TestInsert_Success_AndPublishesEvent(t *testing.T) {
	goal := newValidGoal()
	insertCalled := false
	repo := &mockGoalRepo{
		insertFn: func(ctx context.Context, model *Goal) error {
			insertCalled = true
			return nil
		},
	}
	cacheMock := &mockCache{}
	kb := &mockKeyBuilder{}
	we := &mockWriteExecutor{}
	producer := &mockGoalProducer{}

	svc := newTestService(repo, cacheMock, kb, we, producer)

	err := svc.Insert(context.Background(), goal)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !insertCalled {
		t.Error("Insert was not called")
	}
	if producer.createdEvent == nil {
		t.Fatal("Expected PublishGoalCreated to be called")
	}
	if producer.createdEvent.ID != goal.ID || producer.createdEvent.Name != goal.Name {
		t.Errorf("Event mismatch: got ID=%s Name=%s, want ID=%s Name=%s",
			producer.createdEvent.ID, producer.createdEvent.Name, goal.ID, goal.Name)
	}
}

func TestUpdate_ValidationFailure(t *testing.T) {
	goal := newValidGoal()
	goal.TargetAmount = 0

	updateCalled := false
	repo := &mockGoalRepo{
		updateFn: func(ctx context.Context, model *Goal) error {
			updateCalled = true
			return nil
		},
	}
	cacheMock := &mockCache{}
	kb := &mockKeyBuilder{}
	we := &mockWriteExecutor{}
	producer := &mockGoalProducer{}

	svc := newTestService(repo, cacheMock, kb, we, producer)

	err := svc.Update(context.Background(), goal)

	if err == nil {
		t.Fatal("Expected validation error")
	}
	if updateCalled {
		t.Error("Repo Update should NOT be called for invalid goal")
	}
}

func TestUpdate_Success(t *testing.T) {
	goal := newValidGoal()
	updateCalled := false
	repo := &mockGoalRepo{
		updateFn: func(ctx context.Context, model *Goal) error {
			updateCalled = true
			return nil
		},
	}
	cacheMock := &mockCache{}
	kb := &mockKeyBuilder{}
	we := &mockWriteExecutor{}
	producer := &mockGoalProducer{}

	svc := newTestService(repo, cacheMock, kb, we, producer)

	err := svc.Update(context.Background(), goal)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !updateCalled {
		t.Error("Update was not called")
	}
}

func TestDeleteById_Success_AndPublishesEvent(t *testing.T) {
	id := uuid.New()
	deleteCalled := false
	repo := &mockGoalRepo{
		deleteFn: func(ctx context.Context, goalId uuid.UUID) error {
			deleteCalled = true
			return nil
		},
	}
	cacheMock := &mockCache{}
	kb := &mockKeyBuilder{}
	we := &mockWriteExecutor{}
	producer := &mockGoalProducer{}

	svc := newTestService(repo, cacheMock, kb, we, producer)

	err := svc.DeleteById(context.Background(), id)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !deleteCalled {
		t.Error("DeleteById was not called")
	}
	if producer.deletedEvent == nil {
		t.Fatal("Expected PublishGoalDeleted to be called")
	}
	if producer.deletedEvent.ID != id {
		t.Errorf("Event ID mismatch: got %s, want %s", producer.deletedEvent.ID, id)
	}
}

func TestDeleteById_Error(t *testing.T) {
	repo := &mockGoalRepo{
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			return apiError.ErrRecordNotFound
		},
	}
	cacheMock := &mockCache{}
	kb := &mockKeyBuilder{}
	we := &mockWriteExecutor{}
	producer := &mockGoalProducer{}

	svc := newTestService(repo, cacheMock, kb, we, producer)

	err := svc.DeleteById(context.Background(), uuid.New())

	if !errors.Is(err, apiError.ErrRecordNotFound) {
		t.Errorf("Expected ErrRecordNotFound, got %v", err)
	}
	if producer.deletedEvent != nil {
		t.Error("PublishGoalDeleted should NOT be called on error")
	}
}
