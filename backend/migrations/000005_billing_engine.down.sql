-- 000005_billing_engine.down.sql

DROP TRIGGER IF EXISTS update_folios_updated_at ON folios;
DROP FUNCTION IF EXISTS generate_invoice_number;
DROP FUNCTION IF EXISTS generate_folio_number;
DROP SEQUENCE IF EXISTS invoice_number_seq;
DROP SEQUENCE IF EXISTS folio_number_seq;
DROP TABLE IF EXISTS check_in_details;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS invoice_line_items;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS line_items;
DROP TABLE IF EXISTS folios;
