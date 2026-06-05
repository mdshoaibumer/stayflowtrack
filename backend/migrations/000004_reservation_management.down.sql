-- 000004_reservation_management.down.sql

DROP TRIGGER IF EXISTS update_reservations_updated_at ON reservations;
DROP FUNCTION IF EXISTS check_reservation_conflict;
DROP TABLE IF EXISTS reservations;
