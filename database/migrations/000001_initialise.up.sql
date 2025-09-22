BEGIN;

CREATE TABLE IF NOT EXISTS outbox (
				railreader_message_id SERIAL PRIMARY KEY
				,body jsonb NOT NULL
);

CREATE TABLE IF NOT EXISTS locations (
				location_id text PRIMARY KEY
				,computerised_reservation_system_id text
				,train_operating_company_id text 
				,name text NOT NULL
);

CREATE TABLE IF NOT EXISTS train_operating_companies (
				train_operating_company_id text PRIMARY KEY
				,name text NOT NULL
				,url text
);

CREATE TABLE IF NOT EXISTS late_reasons (
				late_reason_id int PRIMARY KEY
				,description text NOT NULL
);

CREATE TABLE IF NOT EXISTS cancellation_reasons (
				cancellation_reason_id int PRIMARY KEY
				,description text NOT NULL
);

CREATE TABLE IF NOT EXISTS via_conditions (
				sequence int PRIMARY KEY 
				,display_at_location_id text NOT NULL
				,first_required_location_id text NOT NULL
				,second_required_location_id text
				,destination_required_location_id text NOT NULL
				,text text NOT NULL
);

CREATE TABLE IF NOT EXISTS customer_information_systems (
				customer_information_system_id text PRIMARY KEY
				,name text NOT NULL
);

CREATE TABLE IF NOT EXISTS loading_categories (	
				loading_category_id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY
				,loading_category_code text NOT NULL
				,train_operating_company_id text
				,CONSTRAINT uq_code_toc UNIQUE NULLS NOT DISTINCT (loading_category_code, train_operating_company_id) 
				,name text NOT NULL
				,description_typical text NOT NULL
				,description_expected	text NOT NULL
				,definition text NOT NULL
				,colour text NOT NULL
				,image text NOT NULL
);

CREATE TABLE IF NOT EXISTS message_xml (
				message_id text PRIMARY KEY
				,sequence bigint NOT NULL
				,xml xml NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
				message_id text PRIMARY KEY
				,sent_at timestamp WITH TIME ZONE NOT NULL
				,last_received_at timestamp WITH TIME ZONE NOT NULL
				,version text NOT NULL
);

CREATE TABLE IF NOT EXISTS message_response (
				message_id text PRIMARY KEY
				,snapshot boolean NOT NULL
				,source text
				,source_system text
				,request_id text
);

CREATE TABLE IF NOT EXISTS timetables (
				timetable_id text PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS alarms (
		  	alarm_id int PRIMARY KEY
				,has_cleared boolean NOT NULL
				,train_describer_failure text -- a specific train describer that is suspected to have failed
				,all_train_describers_failed boolean
				,tyrell_failed boolean
);

CREATE TABLE IF NOT EXISTS schedules (
				schedule_id text PRIMARY KEY -- this is the RID, renamed to be consistent with other tables

				-- TrainIdentifiers
				,uid text NOT NULL
				,scheduled_start_date date NOT NULL

				-- Schedule
				,headcode text NOT NULL
				,retail_service_id text
				,train_operating_company_id text NOT NULL -- no foreign key contraint here, reference data on TOCs is not complete
				,service text NOT NULL
				,category text NOT NULL
				,is_passenger_service boolean NOT NULL
				,is_active boolean NOT NULL
				,is_deleted boolean NOT NULL
				,is_charter boolean NOT NULL

				,cancellation_reason_id int
				,cancellation_reason_location_id text
				,cancellation_reason_is_near_location boolean

				,late_reason_id int
				,late_reason_location_id text
				,late_reason_is_near_location boolean

				,diverted_via_location_id text

				-- Journey
				,is_cancelled boolean NOT NULL
);

CREATE TABLE IF NOT EXISTS schedules_messages (
				schedule_id text
				,message_id text
				,PRIMARY KEY (schedule_id, message_id)
);

CREATE TABLE IF NOT EXISTS schedules_timetables (
				schedule_id text
				,timetable_id text
				,PRIMARY KEY (schedule_id, timetable_id)
);

CREATE TABLE IF NOT EXISTS schedules_locations (
				schedule_id text
				,sequence int
				,PRIMARY KEY (schedule_id, sequence)

				-- Schedule
				,location_id text NOT NULL
				,activities text ARRAY
				,planned_activities text ARRAY
				,is_cancelled boolean NOT NULL
				,formation_id text
				,is_affected_by_diversion boolean NOT NULL

				,type text NOT NULL
				,public_arrival_time timestamp WITH TIME ZONE
				,public_departure_time timestamp WITH TIME ZONE
				,working_arrival_time timestamp WITH TIME ZONE
				,working_passing_time timestamp WITH TIME ZONE
				,working_departure_time timestamp WITH TIME ZONE
				,routing_delay interval
				,false_destination_location_id text

				,cancellation_reason_id int
				,cancellation_reason_location_id text
				,cancellation_reason_is_near_location boolean

				-- Journey
				,platform text
);

ALTER TABLE schedules_messages ADD CONSTRAINT fk_schedule FOREIGN KEY(schedule_id) REFERENCES schedules(schedule_id) ON DELETE CASCADE;
ALTER TABLE schedules_messages ADD CONSTRAINT fk_message FOREIGN KEY(message_id) REFERENCES messages(message_id) ON DELETE CASCADE;
ALTER TABLE schedules_timetables ADD CONSTRAINT fk_schedule FOREIGN KEY(schedule_id) REFERENCES schedules(schedule_id) ON DELETE CASCADE;
ALTER TABLE schedules_timetables ADD CONSTRAINT fk_timetable FOREIGN KEY(timetable_id) REFERENCES timetables(timetable_id) ON DELETE CASCADE;
ALTER TABLE schedules ADD CONSTRAINT fk_cancellation_reason_location FOREIGN KEY(cancellation_reason_location_id) REFERENCES locations(location_id);
ALTER TABLE schedules ADD CONSTRAINT fk_late_reason_location FOREIGN KEY(late_reason_location_id) REFERENCES locations(location_id);
ALTER TABLE schedules ADD CONSTRAINT fk_diverted_via_location FOREIGN KEY(diverted_via_location_id) REFERENCES locations(location_id);
ALTER TABLE schedules ADD CONSTRAINT fk_cancellation_reason FOREIGN KEY(cancellation_reason_id) REFERENCES cancellation_reasons(cancellation_reason_id);
ALTER TABLE schedules ADD CONSTRAINT fk_late_reason FOREIGN KEY(late_reason_id) REFERENCES late_reasons(late_reason_id);
ALTER TABLE schedules_locations ADD CONSTRAINT fk_schedule FOREIGN KEY(schedule_id) REFERENCES schedules(schedule_id) ON DELETE CASCADE;
ALTER TABLE schedules_locations ADD CONSTRAINT fk_cancellation_reason FOREIGN KEY(cancellation_reason_id) REFERENCES cancellation_reasons(cancellation_reason_id);
ALTER TABLE schedules_locations ADD CONSTRAINT fk_location FOREIGN KEY(location_id) REFERENCES locations(location_id);
ALTER TABLE schedules_locations ADD CONSTRAINT fk_false_destination_location FOREIGN KEY(false_destination_location_id) REFERENCES locations(location_id);
ALTER TABLE schedules_locations ADD CONSTRAINT fk_cancellation_reason_location FOREIGN KEY(cancellation_reason_location_id) REFERENCES locations(location_id);

COMMIT;
