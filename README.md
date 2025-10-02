![Rail Reader logo banner](branding/wide_banner.png)
# railreader

Self-hostable middleman between UK rail datafeeds and your project.

RailReader consumes realtime and static data from multiple railway datasources, and stores accumulated data in a database.
This data can then be queried or subscribed to via multiple outputs.

The aim of RailReader is to provide useful and modern APIs for handling data from the UK's railway network by taking existing data feeds and transforming them into formats that are easier to work with for developers who want to save time.

> [!WARNING]
> This project is very much in **alpha**.
> Data produced may not be fully accurate yet!
>
> The database schema will continue to change without migrations until a first version is released.
> Be prepared to drop your database when updating.

## Inputs
- (**work-in-progress**) Darwin Real Time Train Information XML Push Port (Rail Delivery Group)
- Darwin Timetable Files (Rail Delivery Group)

## Outputs
- SQL queries on the database
- (**TODO**) [General Transit Feed Specification](https://gtfs.org/documentation/schedule/reference/)
- (**TODO**) [General Transit Feed Specification Realtime](https://gtfs.org/documentation/realtime/reference/)

## Resources used
|Input name|Schemas|Documentation|
|----------|-------|-------------|
|Darwin|[XML Schema Definition](resources/darwin_push_port_v24_xsd.zip)|[P75301004 Issue 24](resources/P75301004.pdf), [CIF specification version 29](resources/CIF_v29.pdf)|

### Software to display XML schemas

For dealing with very large XML schemas with lots of types split accross multiple files, I've found Altova XMLSpy to do an extremly good job of exploring the whole schema visually. It is paid software, and is Windows only, but a 30-day free trial is availiable without payment details and if you plan to work with any of the XML schemas, it's almost certainly worth the effort to set up.

## Setup instructions

### PostgreSQL

#### Docker Compose

The included Docker Compose file will run a PostgreSQL server on port 5432 with the username `postgres` and the password `change_me`.
```bash
sudo docker compose up
```
You must then create a database on the server yourself using the PostgreSQL CLI:
```bash
sudo docker run -it --rm --network host postgres psql -h localhost -U postgres
```
```sql
CREATE DATABASE railreader;
```
Use this database URL for Docker: `postgres://postgres:change_me@localhost:5432/railreader?sslmode=disable`.

#### NixOS

There is a provided NixOS module to configure and host both PostgreSQL and RailReader, availiable from this repository's flake at under `nixosModules.railreader`.
For configuration options, read [`service.nix`](service.nix).

#### Your own

Any PostgreSQL database hosted by any other manner will work.

## How to get data in:

- ### Darwin Real Time Train Information
    Subscribe to the [Darwin Real Time Train Information](https://raildata.org.uk/dashboard/dataProduct/P-d3bf124c-1058-4040-8a62-87181a877d59/overview) product via the [Rail Data Marketplace](https://raildata.org.uk) and use the Kafka subscription details for the XML topic on the Pub/Sub page.
- ### Darwin Timetable Files
    Use the S3 details under "Darwin File Information" in the [National Rail Data Portal](https://opendata.nationalrail.co.uk/).

## How to get data out:

- ### PostgreSQL
    Write SQL queries to select from the database.

    #### Example queries

    - Destinations of trains originating from Leeds:

        ```sql
        SELECT name FROM locations l JOIN
            (
                SELECT location_id
                    FROM (
                        SELECT s.schedule_id FROM schedules s
                        JOIN schedule_locations sl ON s.schedule_id = sl.schedule_id
                        WHERE
                            sl.location_id = 'LEEDS'
                            AND sl.type = 'OR' -- Originates from Leeds
                            AND sl.working_departure_time >= to_char(now(), 'HH24:MI') -- Departs after now
                            AND s.scheduled_start_date >= now()::date
                        ORDER BY sl.working_departure_time -- List in order of departure
                    ) as ssli
                JOIN schedule_locations sl ON ssli.schedule_id = sl.schedule_id
                    WHERE
                        sl.type = 'DT' -- Get the destination location
            ) as ssli2
        ON ssli2.location_id = l.location_id; -- Use display names over location_ids
        ```

## AI declaration

Github Copilot was used to autocomplete predictable or repetative lines of code.
All logic and documentation (including this file) was written **without** the assistance of AI in order to ensure accuracy.
