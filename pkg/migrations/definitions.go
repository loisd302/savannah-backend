package migrations

import (
	"gorm.io/gorm"
)

// getAllMigrations returns all available migrations in order
func getAllMigrations() []MigrationItem {
	return []MigrationItem{
		{
			Version:     "001_create_customers_table",
			Description: "Create customers table with UUID primary keys",
			Up:          createCustomersTable,
			Down:        dropCustomersTable,
		},
		{
			Version:     "002_create_orders_table",
			Description: "Create orders table with UUID and foreign key to customers",
			Up:          createOrdersTable,
			Down:        dropOrdersTable,
		},
		{
			Version:     "003_create_history_tables",
			Description: "Create audit history tables for customers and orders",
			Up:          createHistoryTables,
			Down:        dropHistoryTables,
		},
		{
			Version:     "004_add_optimized_indexes",
			Description: "Add performance indexes and constraints",
			Up:          addOptimizedIndexes,
			Down:        dropOptimizedIndexes,
		},
		{
			Version:     "005_add_audit_triggers",
			Description: "Add triggers for automatic audit trail",
			Up:          addAuditTriggers,
			Down:        dropAuditTriggers,
		},
	}
}

// Migration 001: Create customers table
func createCustomersTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS customers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(32) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			phone VARCHAR(20),
			email VARCHAR(255),
			version INTEGER DEFAULT 1,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);
	`).Error
}

func dropCustomersTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS customers CASCADE").Error
}

// Migration 002: Create orders table
func createOrdersTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			customer_id UUID NOT NULL,
			item VARCHAR(255) NOT NULL,
			amount NUMERIC(12,2) NOT NULL,
			ordered_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			status VARCHAR(20) DEFAULT 'pending',
			sms_sent_at TIMESTAMPTZ,
			version INTEGER DEFAULT 1,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
		);
	`).Error
}

func dropOrdersTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS orders CASCADE").Error
}

// Migration 003: Create history tables
func createHistoryTables(db *gorm.DB) error {
	// Create customers_history table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS customers_history (
			id UUID NOT NULL,
			code VARCHAR(32) NOT NULL,
			name VARCHAR(255) NOT NULL,
			phone VARCHAR(20),
			email VARCHAR(255),
			version INTEGER,
			valid_from TIMESTAMPTZ NOT NULL,
			valid_to TIMESTAMPTZ,
			changed_by VARCHAR(100),
			PRIMARY KEY (id, version)
		);
	`).Error; err != nil {
		return err
	}

	// Create orders_history table
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS orders_history (
			id UUID NOT NULL,
			customer_id UUID NOT NULL,
			item VARCHAR(255) NOT NULL,
			amount NUMERIC(12,2) NOT NULL,
			ordered_at TIMESTAMPTZ,
			status VARCHAR(20),
			sms_sent_at TIMESTAMPTZ,
			version INTEGER,
			valid_from TIMESTAMPTZ NOT NULL,
			valid_to TIMESTAMPTZ,
			changed_by VARCHAR(100),
			PRIMARY KEY (id, version)
		);
	`).Error
}

func dropHistoryTables(db *gorm.DB) error {
	if err := db.Exec("DROP TABLE IF EXISTS customers_history CASCADE").Error; err != nil {
		return err
	}
	return db.Exec("DROP TABLE IF EXISTS orders_history CASCADE").Error
}

// Migration 004: Add optimized indexes
func addOptimizedIndexes(db *gorm.DB) error {
	// Enable required extensions first
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		return err
	}

	// Customer indexes
	queries := []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_code ON customers(code)",
		"CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone)",
		"CREATE INDEX IF NOT EXISTS idx_customers_active ON customers(is_active) WHERE is_active = TRUE",
		"CREATE INDEX IF NOT EXISTS idx_customers_name_gin ON customers USING gin(name gin_trgm_ops)",
		
		// Order indexes
		"CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)",
		"CREATE INDEX IF NOT EXISTS idx_orders_ordered_at ON orders(ordered_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_orders_customer_status ON orders(customer_id, status) WHERE status = 'pending'",
		"CREATE INDEX IF NOT EXISTS idx_orders_active ON orders(is_active) WHERE is_active = TRUE",
		
		// History table indexes
		"CREATE INDEX IF NOT EXISTS idx_customers_history_valid ON customers_history(id, valid_from, valid_to)",
		"CREATE INDEX IF NOT EXISTS idx_orders_history_valid ON orders_history(id, valid_from, valid_to)",
	}

	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			return err
		}
	}

	return nil
}

func dropOptimizedIndexes(db *gorm.DB) error {
	indexes := []string{
		"idx_customers_code",
		"idx_customers_phone",
		"idx_customers_active",
		"idx_customers_name_gin",
		"idx_orders_customer_id",
		"idx_orders_status",
		"idx_orders_ordered_at",
		"idx_orders_customer_status",
		"idx_orders_active",
		"idx_customers_history_valid",
		"idx_orders_history_valid",
	}

	for _, index := range indexes {
		if err := db.Exec("DROP INDEX IF EXISTS " + index).Error; err != nil {
			return err
		}
	}

	return nil
}

// Migration 005: Add audit triggers
func addAuditTriggers(db *gorm.DB) error {
	// Enable pg_trgm extension for trigram search
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		return err
	}

	// Create audit trigger function
	if err := db.Exec(`
		CREATE OR REPLACE FUNCTION audit_trigger_func()
		RETURNS TRIGGER AS $$
		BEGIN
			IF TG_OP = 'UPDATE' THEN
				-- Insert old version into history
				IF TG_TABLE_NAME = 'customers' THEN
					INSERT INTO customers_history (id, code, name, phone, email, version, valid_from, valid_to, changed_by)
					VALUES (OLD.id, OLD.code, OLD.name, OLD.phone, OLD.email, OLD.version, OLD.updated_at, CURRENT_TIMESTAMP, 'system');
				ELSIF TG_TABLE_NAME = 'orders' THEN
					INSERT INTO orders_history (id, customer_id, item, amount, ordered_at, status, sms_sent_at, version, valid_from, valid_to, changed_by)
					VALUES (OLD.id, OLD.customer_id, OLD.item, OLD.amount, OLD.ordered_at, OLD.status, OLD.sms_sent_at, OLD.version, OLD.updated_at, CURRENT_TIMESTAMP, 'system');
				END IF;
				-- Increment version
				NEW.version = OLD.version + 1;
				NEW.updated_at = CURRENT_TIMESTAMP;
				RETURN NEW;
			ELSIF TG_OP = 'DELETE' THEN
				-- Insert deleted record into history
				IF TG_TABLE_NAME = 'customers' THEN
					INSERT INTO customers_history (id, code, name, phone, email, version, valid_from, valid_to, changed_by)
					VALUES (OLD.id, OLD.code, OLD.name, OLD.phone, OLD.email, OLD.version, OLD.updated_at, CURRENT_TIMESTAMP, 'system');
				ELSIF TG_TABLE_NAME = 'orders' THEN
					INSERT INTO orders_history (id, customer_id, item, amount, ordered_at, status, sms_sent_at, version, valid_from, valid_to, changed_by)
					VALUES (OLD.id, OLD.customer_id, OLD.item, OLD.amount, OLD.ordered_at, OLD.status, OLD.sms_sent_at, OLD.version, OLD.updated_at, CURRENT_TIMESTAMP, 'system');
				END IF;
				RETURN OLD;
			END IF;
			RETURN NULL;
		END;
		$$ LANGUAGE plpgsql;
	`).Error; err != nil {
		return err
	}

	// Create triggers for customers table
	if err := db.Exec(`
		CREATE TRIGGER customers_audit_trigger
			AFTER UPDATE OR DELETE ON customers
			FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();
	`).Error; err != nil {
		return err
	}

	// Create triggers for orders table
	return db.Exec(`
		CREATE TRIGGER orders_audit_trigger
			AFTER UPDATE OR DELETE ON orders
			FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();
	`).Error
}

func dropAuditTriggers(db *gorm.DB) error {
	queries := []string{
		"DROP TRIGGER IF EXISTS customers_audit_trigger ON customers",
		"DROP TRIGGER IF EXISTS orders_audit_trigger ON orders",
		"DROP FUNCTION IF EXISTS audit_trigger_func()",
	}

	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			return err
		}
	}

	return nil
}
