package database

import (
	"flag"
	"fmt"
	"sync"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/XSAM/otelsql"
	"github.com/bloominlabs/baseplate-go/config/env"
	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-multierror"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

type Client interface {
	Close() error
}

// TODO: we can probably put this in baseplate if we can figure out how to make
// ent.Client generic
type DatabaseConfig[T Client] struct {
	sync.RWMutex

	Host     string `toml:"host"`
	Database string `toml:"database"`
	Port     string `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`

	CreateEntClient func(dialect.Driver) T

	client *T
}

// set the database client manually. useful when writing tests with an already
// initialized client.
//
// WARNING: if you use this parameter, be careful to not use CreateClient() as
// it will overwrite the manually set client. I don't currently have a good
// solution to get around this.
func (c *DatabaseConfig[T]) WithClient(client *T) {
	c.client = client
}

func (c *DatabaseConfig[T]) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Host, "database.host", env.GetEnvStrDefault("DATABASE_HOST", ""), "database host to connect to")
	f.StringVar(&c.Database, "database.database", env.GetEnvStrDefault("DATABASE_DATABASE", ""), "the database to connect to")
	f.StringVar(&c.Port, "database.port", env.GetEnvStrDefault("DATABASE_PORT", ""), "the port of the database host to connect to. NOTE: this is currently unused")
	f.StringVar(&c.Username, "database.username", env.GetEnvStrDefault("DATABASE_USERNAME", ""), "the username to authenticate to the database with")
	f.StringVar(&c.Password, "database.password", env.GetEnvStrDefault("DATABASE_PASSWORD", ""), "the password to authenticate to the database with")
}

func (c *DatabaseConfig[T]) Validate() error {
	var validationErrors error
	if c.Host == "" {
		validationErrors = multierror.Append(validationErrors, fmt.Errorf("did not find a 'host'. did you specify '-database.host' or 'DATABASE_HOST'?"))
	}
	if c.Database == "" {
		validationErrors = multierror.Append(validationErrors, fmt.Errorf("did not find a 'database'. did you specify '-database.database' or 'DATABASE_DATABASE'?"))
	}
	if c.Username == "" {
		validationErrors = multierror.Append(validationErrors, fmt.Errorf("did not find a 'username'. did you specify '-database.username' or 'DATABASE_USERNAME'?"))
	}
	if c.Password == "" {
		validationErrors = multierror.Append(validationErrors, fmt.Errorf("did not find a 'password'. did you specify '-database.password' or 'DATABASE_PASSWORD'?"))
	}

	if c.CreateEntClient == nil {
		validationErrors = multierror.Append(validationErrors, fmt.Errorf("did not find a 'CreateEntClient'. did you set the CreateEntClient parameter?"))
	}

	return validationErrors
}

func (c *DatabaseConfig[T]) Merge(o *DatabaseConfig[T]) error {
	c.Host = o.Host
	c.Database = o.Database
	c.Port = o.Port
	c.Username = o.Username
	c.Password = o.Password

	client, err := c.CreateClient()
	if err != nil {
		return err
	} else {
		err := c.Cleanup()
		if err != nil {
			return fmt.Errorf("failed to cleanup previous client: %w", err)
		}
		c.Lock()
		c.client = client
		c.Unlock()
	}

	return nil
}

func (c *DatabaseConfig[T]) CreateClient() (*T, error) {
	params := map[string]string{
		"parseTime": "true",
		"charset":   "utf8mb4",
		"loc":       "UTC",
		"tls":       "true",
	}

	mc := mysql.Config{
		User:                 c.Username,
		Passwd:               c.Password,
		Net:                  "tcp",
		Addr:                 c.Host,
		DBName:               c.Database,
		AllowNativePasswords: true,
		Params:               params,
	}

	db, err := otelsql.Open(dialect.MySQL, mc.FormatDSN(), otelsql.WithAttributes(
		semconv.DBSystemMySQL,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(
		semconv.DBSystemMySQL,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to register db metrics: %w", err)
	}
	if c.CreateEntClient == nil {
		panic("CreateEntClient not specified. please set it manually")
	}
	drv := sql.OpenDB(dialect.MySQL, db)
	client := c.CreateEntClient(drv)
	return &client, nil
}

func (c *DatabaseConfig[T]) GetClient() (T, error) {
	if c.client == nil {
		client, err := c.CreateClient()
		if err != nil {
			return *client, err
		}
		c.Lock()
		c.client = client
		c.Unlock()

		return *client, err
	}

	c.RLock()
	defer c.RUnlock()
	return *c.client, nil
}

func (c *DatabaseConfig[T]) Cleanup() error {
	c.Lock()
	defer c.Unlock()

	if c.client != nil {
		err := (*c.client).Close()
		c.client = nil
		return err
	}

	return nil
}
