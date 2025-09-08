{ config, lib, railreader, ... }:
let
  types = lib.types;
  cfg = config.services.railreader;
in
{
  options.services.railreader = {
    enable = lib.mkEnableOption "Whether to enable the railreader service.";
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
          Path to a private host key file for the SFTP server.
          You can generate a new keypair using:
          ```
            ssh-keygen -t ed25519 -f host_key -N ""
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
        database = {
          name = lib.mkOption {
            type = types.str;
            default = "railreader_darwin";
            description = ''
              Database name to use for storing Darwin data.
              This is also used as a unix user name for database access.
            '';
          };
        };
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
          cfg.ingest.darwin.database.name
        ];
        ensureUsers = [
          {
            name = cfg.ingest.darwin.database.name;
            ensureDBOwnership = true;
          }
        ];
      };
      systemd.targets.railreader-base = {
        description = "Common base for all railreader services.";
        unitConfig = {
          StateDirectory = "railreader";
          StateDirectoryMode = "0700";
        };
      };
      systemd.services.railreader-sftp = let sftpcfg = cfg.sftp; in {
        description = "Railreader SFTP Server";
        requires = [ "railreader-base.target" ];
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
          SFTP_DARWIN_DIRECTORY = "/var/lib/railreader/darwin";
        };
        script = "${railreader}/bin/railreader sftp";
        scriptArgs = "--hashed-password-file=$CREDENTIALS_DIRECTORY/sftpHashedPassword --private-host-key-file=$CREDENTIALS_DIRECTORY/sftpPrivateHostKey";
        serviceConfig = {
          LoadCredential = [ "sftpHashedPassword:${sftpcfg.hashedPasswordFile}" "sftpPrivateHostKey:${sftpcfg.privateHostKeyFile}" ];
        };
      };
      systemd.services.railreader-ingest = let ingcfg = cfg.ingest; in {
        description = "Railreader Ingest";
        requires = [ "railreader-base.target" "postgresql.service" "railreader-sftp.service" ];
        wants = [ "network-online.target" "postgresql.service" "railreader-sftp.service" ];
        after = [ "network-online.target" "postgresql.service" "railreader-sftp.service" ];
        wantedBy = [ "railreader.target" ];
        partOf = [ "railreader.target" ];
        environment = {
          DARWIN_KAFKA_BROKERS = lib.concatStringsSep "," ingcfg.darwin.kafka.brokers;
          DARWIN_KAFKA_TOPIC = ingcfg.darwin.kafka.topic;
          DARWIN_KAFKA_GROUP = ingcfg.darwin.kafka.group;
          DARWIN_KAFKA_CONNECTION_TIMEOUT = "${toString ingcfg.darwin.kafka.connectionTimeout}s";
          DARWIN_POSTGRESQL_URL = "postgresql://${ingcfg.darwin.database.name}@/$(DARWIN_POSTGRESQL_DBNAME)";
          DARWIN_QUEUE_SIZE = toString ingcfg.darwin.queueSize;
        };
        script = "${railreader}/bin/railreader ingest";
        scriptArgs = "--darwin.kafka.username-file=$CREDENTIALS_DIRECTORY/darwinKafkaUsername --darwin.kafka.password-file=$CREDENTIALS_DIRECTORY/darwinKafkaPassword --darwin.postgresql.dbname=${ingcfg.darwin.database.name}";
        serviceConfig = {
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
