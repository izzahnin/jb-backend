-- Migration: 000003_add_order_status_check_constraint.up.sql
-- Purpose: Add CHECK constraint on orders.status to enforce valid status values
-- Valid statuses: pending, pickup, in_transit, delivered, cancelled

ALTER TABLE orders
ADD CONSTRAINT chk_order_status CHECK (status IN ('pending', 'pickup', 'in_transit', 'delivered', 'cancelled'));
