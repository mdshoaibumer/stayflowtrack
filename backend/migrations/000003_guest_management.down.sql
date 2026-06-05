-- 000003_guest_management.down.sql

DROP TRIGGER IF EXISTS update_guests_updated_at ON guests;
DROP TABLE IF EXISTS guest_documents;
DROP TABLE IF EXISTS guests;
