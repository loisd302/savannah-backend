package repositories

import (
	"backend/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(customer *models.Customer) error {
	return r.db.Create(customer).Error
}

func (r *CustomerRepository) GetByID(id uuid.UUID) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) GetByCode(code string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("code = ? AND is_active = ?", code, true).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) List(query *models.ListCustomersQuery) ([]models.Customer, int64, error) {
	var customers []models.Customer
	var total int64

	db := r.db.Model(&models.Customer{}).Where("is_active = ?", true)

	// Apply filters
	if query.Code != "" {
		db = db.Where("code ILIKE ?", "%"+query.Code+"%")
	}
	if query.Name != "" {
		db = db.Where("name ILIKE ?", "%"+query.Name+"%")
	}
	if query.Phone != "" {
		db = db.Where("phone ILIKE ?", "%"+query.Phone+"%")
	}

	// Get total count
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	err := db.Order("created_at DESC").Find(&customers).Error
	return customers, total, err
}

func (r *CustomerRepository) Update(customer *models.Customer) error {
	return r.db.Save(customer).Error
}

func (r *CustomerRepository) Delete(id uuid.UUID) error {
	// Soft delete by setting is_active = false
	return r.db.Model(&models.Customer{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *CustomerRepository) Exists(code string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Customer{}).Where("code = ? AND is_active = ?", code, true).Count(&count).Error
	return count > 0, err
}