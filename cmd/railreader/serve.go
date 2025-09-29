package main

type ServeCommand struct {
	Host        string `env:"SERVE_HOST" required:"" default:"0.0.0.0:8080" help:"Host to bind the server to."`
	DatabaseURL string `env:"SERVE_POSTGRESQL_URL" required:"" help:"PostgreSQL database URL to read data from."`
}

// TODO: write HTTP server
func (s *ServeCommand) Run() error {
	return nil
}
