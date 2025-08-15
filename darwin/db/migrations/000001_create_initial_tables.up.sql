BEGIN;

CREATE TABLE IF NOT EXISTS train_operating_companies (
				train_operating_company_id text PRIMARY KEY,
				name text NOT NULL,
				url text NOT NULL
);

CREATE TABLE IF NOT EXISTS customer_information_system_sources (
				customer_information_system_source_id text PRIMARY KEY,
				name text NOT NULL
);

CREATE TABLE IF NOT EXISTS late_reasons (
				late_reason_id int PRIMARY KEY,
				description text NOT NULL
);

CREATE TABLE IF NOT EXISTS cancellation_reasons (
				cancellation_reason_id int PRIMARY KEY,
				description text NOT NULL
);

CREATE TABLE IF NOT EXISTS locations (
				location_id text PRIMARY KEY,
				computerised_reservation_system_id text NOT NULL,
				train_operating_company_id text NOT NULL,
				name text NOT NULL
);

CREATE TABLE IF NOT EXISTS schedules (
				schedule_id text PRIMARY KEY, -- this is the RID, renamed to be consistent with other tables

				-- Response
				last_updated timestamp NOT NULL,
				source text,
				source_system text,

				-- TrainIdentifiers
				uid text NOT NULL,
				scheduled_start_date date NOT NULL,

				-- Schedule
				headcode text NOT NULL,
				retail_service_id text,
				train_operating_company_id text NOT NULL, -- no foreign key contraint here, reference data on TOCs is not complete.
				service text NOT NULL,
				category text NOT NULL,
				is_active boolean NOT NULL,
				is_deleted boolean NOT NULL,
				is_charter boolean NOT NULL,

				cancellation_reason_id int,
				cancellation_reason_location_id text,
				cancellation_reason_is_near_location boolean,

				late_reason_id int,
				late_reason_location_id text,
				late_reason_is_near_location boolean,

				diverted_via_location_id text
);

CREATE TABLE IF NOT EXISTS schedules_locations (
				schedule_id text,
				CONSTRAINT fk_schedule FOREIGN KEY(schedule_id) REFERENCES schedules(schedule_id) ON DELETE CASCADE,
				sequence int,
				PRIMARY KEY (schedule_id, sequence),

				-- Schedule
				location_id text,
				activities text,
				planned_activities text,
				is_cancelled boolean NOT NULL,
				is_affected_by_diversion boolean NOT NULL,

				type text NOT NULL,
				public_arrival_time timestamp,
				public_departure_time timestamp,
				working_arrival_time timestamp,
				working_passing_time timestamp,
				working_departure_time timestamp,
				routing_delay interval,
				false_destination_location_id text,

				cancellation_reason_id int,
				cancellation_reason_location_id text,
				cancellation_reason_is_near_location boolean
);

COMMIT;
