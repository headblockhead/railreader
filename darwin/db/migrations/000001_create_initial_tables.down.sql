BEGIN;

DROP TABLE IF EXISTS services_locations_cancellation_reasons;
DROP TABLE IF EXISTS services_cancellation_reasons;
DROP TABLE IF EXISTS services_diversion_reasons;
DROP TABLE IF EXISTS services_locations;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS cancellation_reasons;
DROP TABLE IF EXISTS late_reasons;
DROP TABLE IF EXISTS train_operating_companies;

COMMIT;
