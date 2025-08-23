BEGIN;

DROP TABLE IF EXISTS schedules_locations;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS message_response;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS message_xml;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS cancellation_reasons;
DROP TABLE IF EXISTS late_reasons;
DROP TABLE IF EXISTS customer_information_system_sources;
DROP TABLE IF EXISTS train_operating_companies;

COMMIT;
