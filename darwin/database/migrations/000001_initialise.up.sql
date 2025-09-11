BEGIN;

CREATE TABLE IF NOT EXISTS outbox (
				railreader_darwin_message_id SERIAL PRIMARY KEY -- unique for messages coming out of railreader's Darwin source.
				,body jsonb NOT NULL
);

CREATE TABLE IF NOT EXISTS train_operating_companies (
				train_operating_company_id text PRIMARY KEY
				,name text NOT NULL
				,url text NOT NULL
);

CREATE TABLE IF NOT EXISTS customer_information_system_sources (
				customer_information_system_source_id text PRIMARY KEY
				,name text NOT NULL
);

CREATE TABLE IF NOT EXISTS late_reasons (
				late_reason_id int PRIMARY KEY
				,description text NOT NULL
);

CREATE TABLE IF NOT EXISTS cancellation_reasons (
				cancellation_reason_id int PRIMARY KEY
				,description text NOT NULL
);

CREATE TABLE IF NOT EXISTS locations (
				location_id text PRIMARY KEY
				,computerised_reservation_system_id text NOT NULL
				,train_operating_company_id text NOT NULL
				,name text NOT NULL
);

CREATE TABLE IF NOT EXISTS message_xml (
				message_id text PRIMARY KEY
				,xml xml NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
				message_id text PRIMARY KEY
				,sent_at timestamp NOT NULL
				,last_received_at timestamp NOT NULL
				,version text NOT NULL
);

CREATE TABLE IF NOT EXISTS message_response (
				message_id text PRIMARY KEY
				,snapshot boolean NOT NULL
				,source text
				,source_system text
				,request_id text
);

CREATE TABLE IF NOT EXISTS alarms (
		  	alarm_id int PRIMARY KEY
				,has_cleared boolean NOT NULL
				,train_describer_failure text -- a specific train describer that is suspected to have failed
				,all_train_describers_failed boolean
				,tyrell_failed boolean
);

CREATE TABLE IF NOT EXISTS schedules (
				schedule_id text PRIMARY KEY, -- this is the RID, renamed to be consistent with other tables

				-- PushPortMessage
				,message_id text NOT NULL

				-- TrainIdentifiers
				,uid text NOT NULL
				,scheduled_start_date date NOT NULL

				-- Schedule
				,headcode text NOT NULL
				,retail_service_id text
				,train_operating_company_id text NOT NULL, -- no foreign key contraint here, reference data on TOCs is not complete
				,service text NOT NULL
				,category text NOT NULL
				,is_passenger_service boolean NOT NULL
				,is_active boolean NOT NULL
				,is_deleted boolean NOT NULL
				,is_charter boolean NOT NULL

				,cancellation_reason_id int
				,cancellation_reason_location_id text
				,CONSTRAINT fk_cancellation_reason_location FOREIGN KEY(cancellation_reason_location_id) REFERENCES locations(location_id)
				,cancellation_reason_is_near_location boolean

				,late_reason_id int
				,late_reason_location_id text
				,CONSTRAINT fk_late_reason_location FOREIGN KEY(late_reason_location_id) REFERENCES locations(location_id)
				,late_reason_is_near_location boolean

				,diverted_via_location_id text
);

CREATE TABLE IF NOT EXISTS schedules_locations (
				,schedule_id text
				,CONSTRAINT fk_schedule FOREIGN KEY(schedule_id) REFERENCES schedules(schedule_id) ON DELETE CASCADE
				,sequence int
				,PRIMARY KEY (schedule_id, sequence)

				-- Schedule
				,location_id text NOT NULL
				,CONSTRAINT fk_location FOREIGN KEY(location_id) REFERENCES locations(location_id)
				,activities text ARRAY
				,planned_activities text ARRAY
				,is_cancelled boolean NOT NULL
				,formation_id text
				,is_affected_by_diversion boolean NOT NULL

				,type text NOT NULL
				,public_arrival_time timestamp
				,public_departure_time timestamp
				,working_arrival_time timestamp
				,working_passing_time timestamp
				,working_departure_time timestamp
				,routing_delay interval
				,false_destination_location_id text
				,CONSTRAINT fk_false_destination_location FOREIGN KEY(false_destination_location_id) REFERENCES locations(location_id)

				,cancellation_reason_id int
				,cancellation_reason_location_id text
				,CONSTRAINT fk_cancellation_reason_location FOREIGN KEY(cancellation_reason_location_id) REFERENCES locations(location_id)
				,cancellation_reason_is_near_location boolean
);

COMMIT;
