BEGIN;

CREATE TABLE IF NOT EXISTS train_operating_companies (
				train_operating_company_id text PRIMARY KEY,
				name text NOT NULL,
				url text
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
				location_id text PRIMARY KEY, -- this is the TIPLOC, renamed to be consistent with other tables
				name text NOT NULL,
				CRS text
);

CREATE TABLE IF NOT EXISTS services (
				--  TrainIdentifiers
				service_id text PRIMARY KEY, -- this is the RID, renamed to be consistent with other tables
				UID text NOT NULL,
				scheduled_start_date date NOT NULL,
				-- Schedule
				headcode text NOT NULL,
				retail_service_id text,
				train_operating_company_id text NOT NULL,
				CONSTRAINT fk_train_operating_company FOREIGN KEY(train_operating_company_id) REFERENCES train_operating_companies(train_operating_company_id) ON DELETE CASCADE,
				service text NOT NULL,
				category text NOT NULL,
				active boolean NOT NULL,
				charter boolean NOT NULL,
				cancellation_reason_id int,
				CONSTRAINT fk_cancellation_reason FOREIGN KEY(cancellation_reason_id) REFERENCES cancellation_reasons(cancellation_reason_id) ON DELETE SET NULL,
				late_reason_id int,
				CONSTRAINT fk_late_reason FOREIGN KEY(late_reason_id) REFERENCES late_reasons(late_reason_id) ON DELETE SET NULL,
				diverted_via_location_id text,
				CONSTRAINT fk_diverted_via_location FOREIGN KEY(diverted_via_location_id) REFERENCES locations(location_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS services_locations (
				service_id text,
				CONSTRAINT fk_service FOREIGN KEY(service_id) REFERENCES services(service_id) ON DELETE CASCADE,
				location_id text,
				CONSTRAINT fk_location FOREIGN KEY(location_id) REFERENCES locations(location_id) ON DELETE CASCADE,
				sequence int,
				PRIMARY KEY (service_id, location_id, sequence),
				-- Schedule
				activities text NOT NULL,
				planned_activities text,
				cancelled boolean NOT NULL,
				affected_by_diversion boolean NOT NULL,
				type text NOT NULL,
				-- The elements which are not null depends on the location type
				public_arrival_time timestamp,
				public_departure_time timestamp,
				working_arrival_time timestamp,
				working_passing_time timestamp,
				working_departure_time timestamp,
				false_destination text,
				CONSTRAINT fk_false_destination FOREIGN KEY(false_destination) REFERENCES locations(location_id) ON DELETE SET NULL,
				routing_delay int
);

CREATE TABLE IF NOT EXISTS services_diversion_reasons (
				service_id text,
				CONSTRAINT fk_service FOREIGN KEY(service_id) REFERENCES services(service_id) ON DELETE CASCADE,
				late_reason_id int NOT NULL,
				CONSTRAINT fk_late_reason FOREIGN KEY(late_reason_id) REFERENCES late_reasons(late_reason_id) ON DELETE CASCADE,
				PRIMARY KEY (service_id, late_reason_id),
				locaion_id text,
				CONSTRAINT fk_location FOREIGN KEY(locaion_id) REFERENCES locations(location_id) ON DELETE SET NULL,
				near_location boolean NOT NULL
);

CREATE TABLE IF NOT EXISTS services_cancellation_reasons (
				service_id text,
				CONSTRAINT fk_service FOREIGN KEY(service_id) REFERENCES services(service_id) ON DELETE CASCADE,
				cancellation_reason_id int NOT NULL,
				CONSTRAINT fk_cancellation_reason FOREIGN KEY(cancellation_reason_id) REFERENCES cancellation_reasons(cancellation_reason_id) ON DELETE CASCADE,
				PRIMARY KEY (service_id, cancellation_reason_id),
				locaion_id text,
				CONSTRAINT fk_location FOREIGN KEY(locaion_id) REFERENCES locations(location_id) ON DELETE SET NULL,
				near_location boolean NOT NULL
);

CREATE TABLE IF NOT EXISTS services_locations_cancellation_reasons (
				service_id text,
				CONSTRAINT fk_service FOREIGN KEY(service_id) REFERENCES services(service_id) ON DELETE CASCADE,
				location_id text,
				CONSTRAINT fk_location FOREIGN KEY(location_id) REFERENCES locations(location_id) ON DELETE CASCADE,
				cancellation_reason_id int NOT NULL,
				CONSTRAINT fk_cancellation_reason FOREIGN KEY(cancellation_reason_id) REFERENCES cancellation_reasons(cancellation_reason_id) ON DELETE CASCADE,
				PRIMARY KEY (service_id, location_id, cancellation_reason_id)
);

COMMIT;
