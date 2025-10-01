![Rail Reader logo banner](./wide_banner.png)
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

### Inputs
- (**work-in-progress**) Darwin Real Time Train Information XML Push Port (Rail Delivery Group)
- Darwin Timetable Files (Rail Delivery Group)

### Outputs
- SQL queries to the database
- (**TODO**) [General Transit Feed Specification](https://gtfs.org/documentation/schedule/reference/)
- (**TODO**) [General Transit Feed Specification Realtime](https://gtfs.org/documentation/realtime/reference/)

## Setup Instructions

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
postgres=# CREATE DATABASE railreader;
```
Use this database URL for Docker: `postgres://postgres:change_me@localhost:5432/railreader?sslmode=disable`.

#### NixOS

There is a provided NixOS module to configure and host both PostgreSQL and RailReader, availiable from this repository's flake at under `nixosModules.railreader`.
For configuration options, read [`service.nix`](service.nix).

#### Your own

Any PostgreSQL database hosted by any other manner will work.

### Inputs

#### [Rail Data Marketplace](https://raildata.org.uk)

##### Darwin Real Time Train Information

Subscribe to the [Darwin Real Time Train Information](https://raildata.org.uk/dashboard/dataProduct/P-d3bf124c-1058-4040-8a62-87181a877d59/overview) product and use the Kafka subscription details for the XML topic on the Pub/Sub page.

#### [National Rail Data Portal](https://opendata.nationalrail.co.uk/)

##### Darwin Timetable Files

Use the S3 details under "Darwin File Information".