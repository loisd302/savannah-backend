package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"savannah-backend/internal/domain"
)

// CustomerRepository is a mock implementation of the CustomerRepository interface
type CustomerRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *CustomerRepository) Create(ctx context.Context, customer *domain.Customer) (*domain.Customer, error) {
	args := m.Called(ctx, customer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

// GetByID mocks the GetByID method
func (m *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

// GetByCode mocks the GetByCode method
func (m *CustomerRepository) GetByCode(ctx context.Context, code string) (*domain.Customer, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

// List mocks the List method
func (m *CustomerRepository) List(ctx context.Context, limit, offset int) ([]*domain.Customer, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Customer), args.Error(1)
}

// Update mocks the Update method
func (m *CustomerRepository) Update(ctx context.Context, id uuid.UUID, customer *domain.Customer) (*domain.Customer, error) {
	args := m.Called(ctx, id, customer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

// Delete mocks the Delete method
func (m *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Count mocks the Count method
func (m *CustomerRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}