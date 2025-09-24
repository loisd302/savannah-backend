package repositories

import (
	"time"

	"backend/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) GetByID(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Customer").Where("id = ? AND is_active = ?", id, true).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) List(query *models.ListOrdersQuery) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	db := r.db.Model(&models.Order{}).Where("is_active = ?", true)

	// Apply filters
	if query.CustomerID != uuid.Nil {
		db = db.Where("customer_id = ?", query.CustomerID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
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

	err := db.Preload("Customer").Order("ordered_at DESC").Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepository) GetByCustomerID(customerID uuid.UUID) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Where("customer_id = ? AND is_active = ?", customerID, true).
		Order("ordered_at DESC").Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepository) UpdateStatus(id uuid.UUID, status string, smsSentAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if smsSentAt != nil {
		updates["sms_sent_at"] = *smsSentAt
	}
	return r.db.Model(&models.Order{}).Where("id = ?", id).Updates(updates).Error
}

func (r *OrderRepository) Delete(id uuid.UUID) error {
	// Soft delete by setting is_active = false
	return r.db.Model(&models.Order{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *OrderRepository) GetPendingSMSOrders() ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Customer").
		Where("status = ? AND sms_sent_at IS NULL AND is_active = ?", "pending", true).
		Find(&orders).Error
	return orders, err
}