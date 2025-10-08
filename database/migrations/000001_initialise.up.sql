BEGIN;

CREATE TABLE IF NOT EXISTS outbox (
				railreader_message_id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY 
				,body jsonb NOT NULL
);

-- Reference data

CREATE TABLE IF NOT EXISTS reference_files (
				reference_file_id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY
				,reference_id text UNIQUE NOT NULL
				,filename text NOT NULL
);

CREATE TABLE IF NOT EXISTS locations (
				location_id text PRIMARY KEY
				,name text NOT NULL
				,computerised_reservation_system_id text
				,train_operating_company_id text 
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
				,display_at_computerised_reservation_system_id text NOT NULL
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
);

-- messagexml

CREATE TABLE IF NOT EXISTS message_xml (
				message_id text PRIMARY KEY
				,xml xml NOT NULL
				,kafka_offset bigint NOT NULL
);

-- pport

CREATE TABLE IF NOT EXISTS messages (
				message_id text PRIMARY KEY
				,version text NOT NULL
				,sent_at timestamp WITH TIME ZONE NOT NULL
				,first_received_at timestamp WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS message_status (
				message_id text PRIMARY KEY
				,code text NOT NULL
				,received_at timestamp WITH TIME ZONE NOT NULL
				,description text
);

CREATE TABLE IF NOT EXISTS message_response (
				message_id text PRIMARY KEY
				,is_snapshot boolean NOT NULL
				,source text
				,source_system text
				,request_id text
);

-- timetable

CREATE TABLE IF NOT EXISTS timetable_files (
				timetable_file_id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY
				,timetable_id text UNIQUE NOT NULL
				,first_received_at timestamp WITH TIME ZONE NOT NULL
				,filename text NOT NULL
);

-- associations

CREATE TABLE IF NOT EXISTS associations (
				association_id SERIAL PRIMARY KEY

				-- sources
				,message_id text
				,timetable_id text

				-- Association
				,category text NOT NULL
				,is_cancelled boolean NOT NULL
				,is_deleted boolean NOT NULL
				,main_schedule_id text NOT NULL
				,main_schedule_location_sequence int NOT NULL
				,associated_schedule_id text NOT NULL
				,associated_schedule_location_sequence int NOT NULL
);

-- alarms

CREATE TABLE IF NOT EXISTS alarms (
		  	alarm_id int PRIMARY KEY
				,has_cleared boolean NOT NULL
				,received_at timestamp WITH TIME ZONE NOT NULL
				,cleared_at timestamp WITH TIME ZONE
				,train_describer_failure text -- a specific train describer that is suspected to have failed
				,all_train_describers_failed boolean
				,tyrell_failed boolean
);

-- schedule

CREATE TABLE IF NOT EXISTS schedules (
				schedule_id text PRIMARY KEY -- this is the RID, renamed to be consistent with other tables

				-- sources
				,message_id text
				,timetable_id text

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

CREATE TABLE IF NOT EXISTS schedule_locations (
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
				,public_arrival_time text 
				,public_departure_time text
				,working_arrival_time 				text
				,working_passing_time text 
				,working_departure_time text 
				,routing_delay int
				,false_destination_location_id text

				,cancellation_reason_id int
				,cancellation_reason_location_id text
				,cancellation_reason_is_near_location boolean

				-- Journey
				,platform text
);

ALTER TABLE schedules ADD CONSTRAINT fk_message FOREIGN KEY(message_id) REFERENCES messages(message_id) ON DELETE CASCADE;
ALTER TABLE schedules ADD CONSTRAINT fk_timetable FOREIGN KEY(timetable_id) REFERENCES timetable_files(timetable_id) ON DELETE CASCADE;
ALTER TABLE schedule_locations ADD CONSTRAINT fk_schedule FOREIGN KEY(schedule_id) REFERENCES schedules(schedule_id) ON DELETE CASCADE;

COMMIT;
