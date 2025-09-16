{ config, pkgs, lib, railreader, ... }:
let
  types = lib.types;
  cfg = config.services.railreader;
in
{
  options.services.railreader = {
    enable = lib.mkEnableOption "Whether to enable the railreader service.";
    database = {
      name = lib.mkOption {
        type = types.str;
        default = "railreader";
        description = ''
          Database name to use for storing railreader data.
          This is also used as a unix user name for database access.
        '';
      };
    };
    ingest = {
      darwin = {
        kafka = {
          brokers = lib.mkOption {
            type = types.listOf types.str;
            default = [ "pkc-z3p1v0.europe-west2.gcp.confluent.cloud:9092" ];
            description = ''
              The list of Kafka broker(s) to connect to.
            '';
          };
          topic = lib.mkOption {
            type = types.str;
            default = "prod-1010-Darwin-Train-Information-Push-Port-IIII2_0-XML";
            description = ''
              Kafka topic to subscribe to for Darwin's XML feed.
            '';
          };
          group = lib.mkOption {
            type = types.str;
          };
          usernameFile = lib.mkOption {
            type = types.path;
            description = ''
              File containing the consumer username in plaintext.
            '';
          };
          passwordFile = lib.mkOption {
            type = types.path;
            description = ''
              File containing the consumer password in plaintext.
            '';
          };
          connectionTimeout = lib.mkOption {
            type = types.int;
            default = 30;
            description = ''
              Timeout in seconds for connecting to the Kafka broker.
            '';
          };
        };
        s3 = {
          bucket = lib.mkOption {
            type = types.str;
            default = "darwin.xmltimetable";
            description = ''
              Darwin File Information S3 bucket.
            '';
          };
          prefix = lib.mkOption {
            type = types.str;
            default = "PPTimetable/";
            description = ''
              Prefix within the S3 bucket to read files from.
            '';
          };
          accessKeyFile = lib.mkOption {
            type = types.path;
            description = ''
              File containing the S3 access key in plaintext.
            '';
          };
          secretKeyFile = lib.mkOption {
            type = types.path;
            description = ''
              File containing the S3 secret key in plaintext.
            '';
          };
          region = lib.mkOption {
            type = types.str;
            default = "eu-west-1";
            description = ''
              AWS region the S3 bucket is located in.
            '';
          };
        };
        queueSize = lib.mkOption {
          type = types.int;
          default = 32;
          description = ''
            The maximum number of incoming messages to queue for processing at once. 
            This does not affect data integrity, but will affect memory usage, bandwidth usage on startup, and how long it will take for the service to stop.
          '';
        };
      };
    };
  };
  config = lib.mkIf (config.services.railreader.enable) (lib.mkMerge [
    {
      services.postgresql = {
        enable = true;
        ensureDatabases = [
          cfg.database.name
        ];
        ensureUsers = [
          {
            name = cfg.database.name;
            ensureDBOwnership = true;
          }
        ];
      };
      systemd.services.railreader-ingest = let ingcfg = cfg.ingest; in {
        description = "Railreader Ingest";
        requires = [ "postgresql.service" ];
        wants = [ "network-online.target" "postgresql.service" ];
        after = [ "network-online.target" "postgresql.service" ];
        wantedBy = [ "railreader.target" ];
        partOf = [ "railreader.target" ];
        environment = {
          LOG_LEVEL = "debug";
          POSTGRESQL_URL = "postgresql:///${cfg.database.name}?host=/run/postgresql&sslmode=disable";
          DARWIN_KAFKA_BROKERS = lib.concatStringsSep "," ingcfg.darwin.kafka.brokers;
          DARWIN_KAFKA_TOPIC = ingcfg.darwin.kafka.topic;
          DARWIN_KAFKA_GROUP = ingcfg.darwin.kafka.group;
          DARWIN_KAFKA_CONNECTION_TIMEOUT = "${toString ingcfg.darwin.kafka.connectionTimeout}s";
          DARWIN_S3_BUCKET = ingcfg.darwin.s3.bucket;
          DARWIN_S3_PREFIX = ingcfg.darwin.s3.prefix;
          DARWIN_S3_REGION = ingcfg.darwin.s3.region;
          DARWIN_QUEUE_SIZE = toString ingcfg.darwin.queueSize;
        };
        script = ''
          export DARWIN_KAFKA_USERNAME=$(${pkgs.systemd}/bin/systemd-creds cat darwinKafkaUsername)
          export DARWIN_KAFKA_PASSWORD=$(${pkgs.systemd}/bin/systemd-creds cat darwinKafkaPassword)
          export DARWIN_S3_ACCESS_KEY=$(${pkgs.systemd}/bin/systemd-creds cat darwinS3AccessKey)
          export DARWIN_S3_SECRET_KEY=$(${pkgs.systemd}/bin/systemd-creds cat darwinS3SecretKey)
          ${railreader}/bin/railreader ingest 
        '';
        serviceConfig = {
          DynamicUser = true;
          User = cfg.database.name;
          LoadCredential = [
            "darwinKafkaUsername:${cfg.ingest.darwin.kafka.usernameFile}"
            "darwinKafkaPassword:${cfg.ingest.darwin.kafka.passwordFile}"
            "darwinS3AccessKey:${cfg.ingest.darwin.s3.accessKeyFile}"
            "darwinS3SecretKey:${cfg.ingest.darwin.s3.secretKeyFile}"
          ];
        };
      };
      systemd.targets.railreader = {
        description = "Common target for all railreader services.";
        wantedBy = [ "multi-user.target" ];
      };
    }
  ]);
}
