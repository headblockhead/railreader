BEGIN;

DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS schedules_locations;

DROP TABLE IF EXISTS alarms;
DROP TABLE IF EXISTS timetable_files;

DROP TABLE IF EXISTS message_response;
DROP TABLE IF EXISTS message_status;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS message_xml;

DROP TABLE IF EXISTS loading_categories;
DROP TABLE IF EXISTS customer_information_systems;
DROP TABLE IF EXISTS via_conditions;
DROP TABLE IF EXISTS cancellation_reasons;
DROP TABLE IF EXISTS late_reasons;
DROP TABLE IF EXISTS train_operating_companies;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS reference_files;

DROP TABLE IF EXISTS outbox;

COMMIT;
