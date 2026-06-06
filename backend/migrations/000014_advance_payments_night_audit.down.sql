ALTER TABLE reservations DROP COLUMN IF EXISTS advance_amount;
ALTER TABLE reservations DROP COLUMN IF EXISTS advance_method;
ALTER TABLE reservations DROP COLUMN IF EXISTS advance_reference;
DROP TABLE IF EXISTS night_audits;
