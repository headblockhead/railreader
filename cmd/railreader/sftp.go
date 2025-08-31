package main

type SFTPCommand struct {
	Address  string `group:"Darwin Push Port client:" env:"DARWIN_SFTP_ADDRESS" help:"SFTP server address to listen on for file transfers" default:"127.0.0.1:8022"`
	User     string `group:"Darwin Push Port client:" env:"DARWIN_SFTP_USER" help:"Username for SFTP connections" default:"darwin"`
	Password string `group:"Darwin Push Port client:" env:"DARWIN_SFTP_PASSWORD" help:"Password for SFTP connections" required:""`
	Logging  struct {
		Level string `enum:"debug,info,warn,error" default:"warn"`
		Type  string `enum:"json,console" default:"json"`
	} `embed:"" prefix:"logging."`
}
