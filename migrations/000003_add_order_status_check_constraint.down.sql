-- Migration: 000003_add_order_status_check_constraint.down.sql
-- Purpose: Rollback - remove CHECK constraint

ALTER TABLE orders
DROP CONSTRAINT chk_order_status;
