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
    sftp = {
      hashedPasswordFile = lib.mkOption {
        type = types.path;
        description = ''
          Path to a bcrypt hashed password file for SFTP authentication.
          You can generate a password hash using:
          ```
            nix run nixpkgs#mkpasswd -- -m bcrypt
          ```
        '';
      };
      privateHostKeyFile = lib.mkOption {
        type = types.path;
        description = ''
          Path to a private RSA host key file for the SFTP server.
          You can generate a new keypair using:
          ```
            ssh-keygen -t rsa -f host_key -N ""
          ```
        '';
      };
      listenAddresses = lib.mkOption {
        type = types.listOf (types.submodule {
          options = {
            host = lib.mkOption {
              type = types.nullOr types.str;
              default = null;
              description = ''
                Host, IPv4, or IPv6 address to listen on.
              '';
            };
            port = lib.mkOption {
              type = types.nullOr types.int;
              default = null;
              description = ''
                Port to listen on.
              '';
            };
          };
        });
        default = [
          {
            host = "0.0.0.0";
            port = 64022;
          }
        ];
        description = ''
          List of addresses to listen on.
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
      systemd.services.railreader-sftp = let sftpcfg = cfg.sftp; in {
        description = "Railreader SFTP Server";
        wants = [ "network-online.target" ];
        after = [ "network-online.target" ];
        wantedBy = [ "railreader.target" ];
        partOf = [ "railreader.target" ];
        environment = {
          SFTP_ADDRESSES = lib.strings.concatMapStringsSep ","
            (addr:
              let
                h = if addr.host == null then "" else addr.host;
                p = if addr.port == null then "" else toString addr.port;
              in
              if h == "" && p == "" then ""
              else if h == "" then ":" + p
              else if p == "" then h
              else h + ":" + p
            )
            sftpcfg.listenAddresses;
          SFTP_DARWIN_DIRECTORY = "/var/lib/railreader/sftp/darwin";
        };
        script = ''
          export SFTP_HASHED_PASSWORD=$(${pkgs.systemd}/bin/systemd-creds cat sftpHashedPassword)
          ${railreader}/bin/railreader sftp --private-host-key-file=$CREDENTIALS_DIRECTORY/sftpPrivateHostKey
        '';
        serviceConfig = {
          DynamicUser = true;
          User = cfg.database.name;
          ExecStartPre = ''
            ${pkgs.coreutils}/bin/mkdir -p $SFTP_DARWIN_DIRECTORY
          '';
          LoadCredential = [ "sftpHashedPassword:${sftpcfg.hashedPasswordFile}" "sftpPrivateHostKey:${sftpcfg.privateHostKeyFile}" ];
          StateDirectory = "railreader";
          StateDirectoryMode = "0700";
        };
      };
      systemd.services.railreader-ingest = let ingcfg = cfg.ingest; in {
        description = "Railreader Ingest";
        requires = [ "postgresql.service" "railreader-sftp.service" ];
        wants = [ "network-online.target" "postgresql.service" "railreader-sftp.service" ];
        after = [ "network-online.target" "postgresql.service" "railreader-sftp.service" ];
        wantedBy = [ "railreader.target" ];
        partOf = [ "railreader.target" ];
        environment = {
          POSTGRESQL_URL = "postgresql:///${cfg.database.name}?host=/run/postgresql&sslmode=disable";
          DARWIN_KAFKA_BROKERS = lib.concatStringsSep "," ingcfg.darwin.kafka.brokers;
          DARWIN_KAFKA_TOPIC = ingcfg.darwin.kafka.topic;
          DARWIN_KAFKA_GROUP = ingcfg.darwin.kafka.group;
          DARWIN_KAFKA_CONNECTION_TIMEOUT = "${toString ingcfg.darwin.kafka.connectionTimeout}s";
          DARWIN_QUEUE_SIZE = toString ingcfg.darwin.queueSize;
        };
        script = ''
          export DARWIN_KAFKA_USERNAME=$(${pkgs.systemd}/bin/systemd-creds cat darwinKafkaUsername)
          export DARWIN_KAFKA_PASSWORD=$(${pkgs.systemd}/bin/systemd-creds cat darwinKafkaPassword)
          ${railreader}/bin/railreader ingest 
        '';
        serviceConfig = {
          DynamicUser = true;
          User = cfg.database.name;
          LoadCredential = [ "darwinKafkaUsername:${cfg.ingest.darwin.kafka.usernameFile}" "darwinKafkaPassword:${cfg.ingest.darwin.kafka.passwordFile}" ];
        };
      };
      systemd.targets.railreader = {
        description = "Common target for all railreader services.";
        wantedBy = [ "multi-user.target" ];
      };
    }
  ]);
}
