-- 000002_property_management.down.sql

DROP TRIGGER IF EXISTS update_units_updated_at ON units;
DROP TRIGGER IF EXISTS update_unit_types_updated_at ON unit_types;
DROP TRIGGER IF EXISTS update_properties_updated_at ON properties;

DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS unit_types;
DROP TABLE IF EXISTS properties;
