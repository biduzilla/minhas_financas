package user

import (
	"context"
	"errors"
	"ms_auth/internal/core/domain/apiError"
	"ms_auth/internal/core/filters"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockUserRepo struct {
	findAllFn     func(ctx context.Context, search string, f filters.Filters) ([]*User, filters.Metadata, error)
	findByIdFn    func(ctx context.Context, id uuid.UUID) (*User, error)
	findByEmailFn func(ctx context.Context, email string) (*User, error)
	insertFn      func(ctx context.Context, model *User) error
	updateFn      func(ctx context.Context, model *User) error
	deleteByIdFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserRepo) FindAll(ctx context.Context, search string, f filters.Filters) ([]*User, filters.Metadata, error) {
	return m.findAllFn(ctx, search, f)
}
func (m *mockUserRepo) FindById(ctx context.Context, id uuid.UUID) (*User, error) {
	return m.findByIdFn(ctx, id)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	return m.findByEmailFn(ctx, email)
}
func (m *mockUserRepo) Insert(ctx context.Context, model *User) error {
	return m.insertFn(ctx, model)
}
func (m *mockUserRepo) Update(ctx context.Context, model *User) error {
	return m.updateFn(ctx, model)
}
func (m *mockUserRepo) DeleteById(ctx context.Context, id uuid.UUID) error {
	return m.deleteByIdFn(ctx, id)
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

func (m *mockKeyBuilder) BuildItemKey(id string) string     { return "user:" + id }
func (m *mockKeyBuilder) BuildListKey(params ...any) string { return "user:list" }
func (m *mockKeyBuilder) GetPrefix() string                 { return "user:" }

type mockWriteExecutor struct {
	executeFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockWriteExecutor) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.executeFn != nil {
		return m.executeFn(ctx, fn)
	}
	return fn(ctx)
}

func newValidUser() *User {
	return &User{
		ID:        uuid.New(),
		Email:     "joao@empresa.com",
		Name:      "João",
		Activated: false,
	}
}

func newValidSignUpDTO() CreateUserDTO {
	return CreateUserDTO{
		Email:    "joao@empresa.com",
		Password: "senha123A",
		Name:     "João",
	}
}

func TestSignUp_Sucess(t *testing.T) {
	req := newValidSignUpDTO()
	repo := &mockUserRepo{
		insertFn: func(ctx context.Context, model *User) error {
			return nil
		},
	}
	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	user, err := svc.SignUp(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user == nil {
		t.Fatal("Expected user to be created")
	}

	if user.Email != req.Email || user.Name != req.Name {
		t.Error("User fields not set correctly")
	}
	if len(user.PasswordHash) == 0 {
		t.Error("Password hash should be generated")
	}
}

func TestSignUp_InvalidPassword_MissingLetter(t *testing.T) {
	req := newValidSignUpDTO()
	req.Password = "12345678"

	insertCalled := false
	repo := &mockUserRepo{
		insertFn: func(ctx context.Context, model *User) error {
			insertCalled = true
			return nil
		},
	}

	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	_, err := svc.SignUp(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error")
	}
	if insertCalled {
		t.Error("Insert should NOT be called for invalid password")
	}
}

func TestSignUp_InvalidPassword_MissingDigit(t *testing.T) {
	req := newValidSignUpDTO()
	req.Password = "abcdefgh"

	repo := &mockUserRepo{
		insertFn: func(ctx context.Context, model *User) error {
			t.Error("Insert should not be called")
			return nil
		},
	}
	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	_, err := svc.SignUp(context.Background(), req)

	if err == nil {
		t.Fatal("Expected validation error")
	}
}

func TestSignUp_ErrorFromInsert(t *testing.T) {
	req := newValidSignUpDTO()
	expectedErr := errors.New("db error")
	repo := &mockUserRepo{
		insertFn: func(ctx context.Context, model *User) error {
			return expectedErr
		},
	}
	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	_, err := svc.SignUp(context.Background(), req)

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected %v, got %v", expectedErr, err)
	}
}

func TestFindAll_Success_FromDB(t *testing.T) {
	expectedUsers := []*User{newValidUser()}
	expectedMeta := filters.Metadata{TotalRecords: 1, CurrentPage: 1, PageSize: 10}
	repo := &mockUserRepo{
		findAllFn: func(ctx context.Context, search string, f filters.Filters) ([]*User, filters.Metadata, error) {
			return expectedUsers, expectedMeta, nil
		},
	}
	cacheMock := &mockCache{
		getFn: func(ctx context.Context, key string, dest any) error {
			return errors.New("cache miss")
		},
	}
	svc := NewService(repo, cacheMock, &mockKeyBuilder{}, &mockWriteExecutor{})

	users, meta, err := svc.FindAll(context.Background(), "joao", filters.Filters{Page: 1, PageSize: 10, Sort: "name"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
	if meta.TotalRecords != 1 {
		t.Errorf("Expected TotalRecords 1, got %d", meta.TotalRecords)
	}
}

func TestFindAll_Error(t *testing.T) {
	repo := &mockUserRepo{
		findAllFn: func(ctx context.Context, search string, f filters.Filters) ([]*User, filters.Metadata, error) {
			return nil, filters.Metadata{}, errors.New("query failed")
		},
	}
	cacheMock := &mockCache{getFn: func(ctx context.Context, key string, dest any) error { return errors.New("miss") }}
	svc := NewService(repo, cacheMock, &mockKeyBuilder{}, &mockWriteExecutor{})

	_, _, err := svc.FindAll(context.Background(), "", filters.Filters{Page: 1, PageSize: 10, Sort: "name"})

	if err == nil {
		t.Fatal("Expected error")
	}
}

// ---------- FindById ----------

func TestFindById_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		findByIdFn: func(ctx context.Context, id uuid.UUID) (*User, error) {
			return nil, apiError.ErrRecordNotFound
		},
	}
	cacheMock := &mockCache{getFn: func(ctx context.Context, key string, dest any) error { return errors.New("miss") }}
	svc := NewService(repo, cacheMock, &mockKeyBuilder{}, &mockWriteExecutor{})

	_, err := svc.FindById(context.Background(), uuid.New())

	if !errors.Is(err, apiError.ErrRecordNotFound) {
		t.Errorf("Expected ErrRecordNotFound, got %v", err)
	}
}

func TestFindById_Success(t *testing.T) {
	expected := newValidUser()
	repo := &mockUserRepo{
		findByIdFn: func(ctx context.Context, id uuid.UUID) (*User, error) {
			return expected, nil
		},
	}
	cacheMock := &mockCache{getFn: func(ctx context.Context, key string, dest any) error { return errors.New("miss") }}
	svc := NewService(repo, cacheMock, &mockKeyBuilder{}, &mockWriteExecutor{})

	result, err := svc.FindById(context.Background(), expected.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.ID != expected.ID {
		t.Error("Wrong ID returned")
	}
}

// ---------- Insert (via WriteExecutor) ----------

func TestInsert_ValidationFailure(t *testing.T) {
	invalidModel := &User{Name: "", Email: "invalid"}
	insertCalled := false
	repo := &mockUserRepo{
		insertFn: func(ctx context.Context, model *User) error {
			insertCalled = true
			return nil
		},
	}
	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	err := svc.Insert(context.Background(), invalidModel)

	if err == nil {
		t.Fatal("Expected validation error")
	}
	if insertCalled {
		t.Error("Repository insert should NOT be called for invalid user")
	}
}

func TestInsert_Success(t *testing.T) {
	validModel := newValidUser()
	insertCalled := false
	repo := &mockUserRepo{
		insertFn: func(ctx context.Context, model *User) error {
			insertCalled = true
			return nil
		},
	}
	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	err := svc.Insert(context.Background(), validModel)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !insertCalled {
		t.Error("Repository insert was not called")
	}
}

func TestUpdate_ValidationFailure(t *testing.T) {
	invalidModel := &User{Name: ""}
	updateCalled := false
	repo := &mockUserRepo{
		updateFn: func(ctx context.Context, model *User) error {
			updateCalled = true
			return nil
		},
	}
	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	err := svc.Update(context.Background(), invalidModel)

	if err == nil {
		t.Fatal("Expected validation error")
	}
	if updateCalled {
		t.Error("Update should not be called for invalid user")
	}
}

func TestUpdate_Success(t *testing.T) {
	validModel := newValidUser()
	updateCalled := false
	repo := &mockUserRepo{
		updateFn: func(ctx context.Context, model *User) error {
			updateCalled = true
			return nil
		},
	}
	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	err := svc.Update(context.Background(), validModel)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !updateCalled {
		t.Error("Update was not called")
	}
}

func TestUpdate_RepoError(t *testing.T) {
	validModel := newValidUser()
	repo := &mockUserRepo{
		updateFn: func(ctx context.Context, model *User) error {
			return apiError.ErrEditConflict
		},
	}
	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	err := svc.Update(context.Background(), validModel)

	if !errors.Is(err, apiError.ErrEditConflict) {
		t.Errorf("Expected ErrEditConflict, got %v", err)
	}
}

func TestDeleteById_Success(t *testing.T) {
	deleteCalled := false
	repo := &mockUserRepo{
		deleteByIdFn: func(ctx context.Context, id uuid.UUID) error {
			deleteCalled = true
			return nil
		},
	}
	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	err := svc.DeleteById(context.Background(), uuid.New())

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !deleteCalled {
		t.Error("DeleteById was not called")
	}
}

func TestDeleteById_RepoError(t *testing.T) {
	repo := &mockUserRepo{
		deleteByIdFn: func(ctx context.Context, id uuid.UUID) error {
			return apiError.ErrRecordNotFound
		},
	}
	svc := NewService(repo, &mockCache{}, &mockKeyBuilder{}, &mockWriteExecutor{})

	err := svc.DeleteById(context.Background(), uuid.New())

	if !errors.Is(err, apiError.ErrRecordNotFound) {
		t.Errorf("Expected ErrRecordNotFound, got %v", err)
	}
}
