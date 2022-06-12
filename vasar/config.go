package vasar

import "github.com/df-mc/dragonfly/server"

// Config is an extension of the Dragonfly server config to include fields specific to Vasar.
type Config struct {
	server.Config
	// Vasar contains fields specific to Vasar.
	Vasar struct {
		// Tebex is the Tebex API key.
		Tebex string
		// Whitelisted is true if the server is whitelisted.
		Whitelisted bool
		// Season is the current season of the server.
		Season int
		// Start is the date the season started.
		Start string
		// End is the date the season ends.
		End string
	}
	// Pack contains fields related to the pack.
	Pack struct {
		// Key is the pack encryption key.
		Key string
		// Path is the path to the pack.
		Path string
	}
	// Oomph contains fields specific to Oomph.
	Oomph struct {
		// Address is the address to run Oomph on.
		Address string
	}
	// Sentry contains fields used for Sentry.
	Sentry struct {
		// Release is the release name.
		Release string
		// Dsn is the Sentry Dsn.
		Dsn string
	}
}

// DefaultConfig returns a default config for the server.
func DefaultConfig() Config {
	c := Config{}
	c.Config = server.DefaultConfig()
	c.Vasar.Whitelisted = true
	c.Vasar.Season = 1
	c.Vasar.Start = "Edit this!"
	c.Vasar.End = "Edit this!"
	return c
}
