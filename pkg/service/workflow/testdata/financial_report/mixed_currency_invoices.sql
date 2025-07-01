-- Test data for mixed currency invoices
-- Tests that ConversionAmount is used over Total when available

INSERT INTO invoices (id, number, email, total, subtotal, tax, discount, status, created_at, updated_at, paid_at, conversion_amount)
VALUES 
-- Invoice with ConversionAmount - should use ConversionAmount (100,000 VND)
('invoice-001', 'INV-2025-001', 'client1@example.com', 4.0, 4.0, 0, 0, 'paid', '2025-06-15 10:00:00', '2025-06-15 10:00:00', '2025-06-15 10:00:00', 100000.0),
-- Invoice with no ConversionAmount - should fallback to Total (50,000 VND equivalent)
('invoice-002', 'INV-2025-002', 'client2@example.com', 50000.0, 50000.0, 0, 0, 'paid', '2025-06-20 10:00:00', '2025-06-20 10:00:00', '2025-06-20 10:00:00', 0),
-- Invoice with zero ConversionAmount - should fallback to Total (this case shouldn't happen in practice)
('invoice-003', 'INV-2025-003', 'client3@example.com', 25000.0, 25000.0, 0, 0, 'paid', '2025-06-25 10:00:00', '2025-06-25 10:00:00', '2025-06-25 10:00:00', 0);

-- Expected total: 100,000 + 50,000 = 150,000 VND (invoice-003 should use Total since ConversionAmount is 0)