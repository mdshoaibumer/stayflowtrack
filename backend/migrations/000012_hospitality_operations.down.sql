-- 000012_hospitality_operations.down.sql

DROP TABLE IF EXISTS deposits;
DROP TABLE IF EXISTS room_moves;
DROP TABLE IF EXISTS stay_extensions;
DROP TABLE IF EXISTS maintenance_blocks;
DROP TABLE IF EXISTS corporate_accounts;

ALTER TABLE reservations DROP COLUMN IF EXISTS is_no_show;
ALTER TABLE reservations DROP COLUMN IF EXISTS no_show_at;
ALTER TABLE reservations DROP COLUMN IF EXISTS no_show_charge;
ALTER TABLE reservations DROP COLUMN IF EXISTS corporate_account_id;
