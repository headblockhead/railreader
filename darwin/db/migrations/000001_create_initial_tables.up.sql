BEGIN;

-- SCHEDULES

DO $$ BEGIN
CREATE TYPE disruption_reason AS (
	reason text,
	is_near_tiploc boolean,
	tiploc text
);

CREATE TYPE schedule_location AS (
	location_type text,
	tiploc text,
	activities text,
	planned_activities text,
	cancelled boolean,
	formation_id text,
	is_affected_by_diversion boolean,
	cancellation_reason disruption_reason,
	public_arrival_time time,
	public_departure_time time,
	working_arrival_time time,
	working_departure_time time,
	working_passing_time time,
	routing_delay interval,
	false_destination text
);
EXCEPTION
	WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS schedules (
	rid text PRIMARY KEY,
	uid text NOT NULL,
	start_date date NOT NULL,
	headcode text NOT NULL,
	retail_service_id text,
	train_operating_company text NOT NULL,
	service text NOT NULL,
	category text NOT NULL,
	is_passenger_service boolean NOT NULL,
	is_deactivated boolean NOT NULL,
	is_deleted boolean NOT NULL,
	is_charter boolean NOT NULL,
	cancellation_reason disruption_reason,
	diversion_reason disruption_reason,
	diverted_via text,
	locations schedule_location[] NOT NULL
);

-- ALARMS

CREATE TABLE IF NOT EXISTS alarms (
	alarm_id bigint PRIMARY KEY,
	is_cleared boolean NOT NULL,
	train_describer_failed_area text,
	train_describer_total_feed_failure boolean,
	tyrell_total_feed_failure boolean
);

COMMIT;
