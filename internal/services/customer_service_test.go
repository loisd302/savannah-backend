package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"savannah-backend/internal/domain"
	"savannah-backend/internal/repositories/mocks"
)

type CustomerServiceTestSuite struct {
	suite.Suite
	service    *CustomerService
	mockRepo   *mocks.CustomerRepository
	ctx        context.Context
}

func (suite *CustomerServiceTestSuite) SetupTest() {
	suite.mockRepo = new(mocks.CustomerRepository)
	suite.service = NewCustomerService(suite.mockRepo)
	suite.ctx = context.Background()
}

func (suite *CustomerServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *CustomerServiceTestSuite) TestCreateCustomer_Success() {
	// Arrange
	customerData := &domain.Customer{
		Name:        "John Doe",
		Code:        "CUST001",
		PhoneNumber: "+254700123456",
		Email:       "john@example.com",
	}

	expectedCustomer := &domain.Customer{
		ID:          uuid.New(),
		Name:        "John Doe",
		Code:        "CUST001",
		PhoneNumber: "+254700123456",
		Email:       "john@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	suite.mockRepo.On("Create", suite.ctx, customerData).Return(expectedCustomer, nil)

	// Act
	result, err := suite.service.CreateCustomer(suite.ctx, customerData)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedCustomer.Name, result.Name)
	assert.Equal(suite.T(), expectedCustomer.Code, result.Code)
	assert.Equal(suite.T(), expectedCustomer.Email, result.Email)
	assert.Equal(suite.T(), expectedCustomer.PhoneNumber, result.PhoneNumber)
}

func (suite *CustomerServiceTestSuite) TestCreateCustomer_ValidationError() {
	// Arrange
	invalidCustomer := &domain.Customer{
		Name: "", // Invalid: empty name
		Code: "CUST001",
	}

	// Act
	result, err := suite.service.CreateCustomer(suite.ctx, invalidCustomer)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "name is required")
}

func (suite *CustomerServiceTestSuite) TestGetCustomerByID_Success() {
	// Arrange
	customerID := uuid.New()
	expectedCustomer := &domain.Customer{
		ID:          customerID,
		Name:        "Jane Doe",
		Code:        "CUST002",
		PhoneNumber: "+254700654321",
		Email:       "jane@example.com",
	}

	suite.mockRepo.On("GetByID", suite.ctx, customerID).Return(expectedCustomer, nil)

	// Act
	result, err := suite.service.GetCustomerByID(suite.ctx, customerID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedCustomer.ID, result.ID)
	assert.Equal(suite.T(), expectedCustomer.Name, result.Name)
}

func (suite *CustomerServiceTestSuite) TestGetCustomerByID_NotFound() {
	// Arrange
	customerID := uuid.New()
	suite.mockRepo.On("GetByID", suite.ctx, customerID).Return(nil, domain.ErrCustomerNotFound)

	// Act
	result, err := suite.service.GetCustomerByID(suite.ctx, customerID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), domain.ErrCustomerNotFound, err)
}

func (suite *CustomerServiceTestSuite) TestListCustomers_Success() {
	// Arrange
	expectedCustomers := []*domain.Customer{
		{
			ID:   uuid.New(),
			Name: "Customer 1",
			Code: "CUST001",
		},
		{
			ID:   uuid.New(),
			Name: "Customer 2",
			Code: "CUST002",
		},
	}

	suite.mockRepo.On("List", suite.ctx, 10, 0).Return(expectedCustomers, nil)

	// Act
	result, err := suite.service.ListCustomers(suite.ctx, 10, 0)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 2)
}

func (suite *CustomerServiceTestSuite) TestUpdateCustomer_Success() {
	// Arrange
	customerID := uuid.New()
	updateData := &domain.Customer{
		Name:        "Updated Name",
		PhoneNumber: "+254700999888",
	}

	expectedCustomer := &domain.Customer{
		ID:          customerID,
		Name:        "Updated Name",
		Code:        "CUST001",
		PhoneNumber: "+254700999888",
		Email:       "john@example.com",
		UpdatedAt:   time.Now(),
	}

	suite.mockRepo.On("Update", suite.ctx, customerID, updateData).Return(expectedCustomer, nil)

	// Act
	result, err := suite.service.UpdateCustomer(suite.ctx, customerID, updateData)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedCustomer.Name, result.Name)
	assert.Equal(suite.T(), expectedCustomer.PhoneNumber, result.PhoneNumber)
}

func (suite *CustomerServiceTestSuite) TestDeleteCustomer_Success() {
	// Arrange
	customerID := uuid.New()
	suite.mockRepo.On("Delete", suite.ctx, customerID).Return(nil)

	// Act
	err := suite.service.DeleteCustomer(suite.ctx, customerID)

	// Assert
	assert.NoError(suite.T(), err)
}

func TestCustomerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CustomerServiceTestSuite))
}