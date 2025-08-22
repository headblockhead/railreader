package main

type InterpretCommand struct {
	Darwin InterpretDarwinCommand `cmd:"darwin" help:"Interpret a message from Darwin."`
}

type InterpretDarwinCommand struct {
	MessageID string `arg:"" help:"message_id of a message to (re-)interpret. This will be fetched from the database."`
	File      string `arg:"" help:"Path to a file containing a Darwin message to interpret. This takes precedence over providing a message_id."`

	DryRun bool `help:"Do not write the message to the database."`

	Logging struct {
		Level string `enum:"debug,info,warn,error" default:"info"`
		Type  string `enum:"json,console" default:"console"`
	} `embed:"" prefix:"logging."`
	Socket struct {
		// TODO: grab location from SYSTEMD? /var/run/railreader/railreader.sock
		Location string `env:"SOCKET_LOCATION" help:"Path of the socket file to connect to."`
	} `embed:"" prefix:"socket."`
}

// TODO: implement
