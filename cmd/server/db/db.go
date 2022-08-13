package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type store struct {
	DB  *pgx.Conn
	Ctx context.Context
}

func New(ctx context.Context, envCfg *config.Config) (*store, func()) {
	conn, err := pgx.Connect(ctx, envCfg.PgDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	closer := func() {
		conn.Close(ctx)
	}

	return &store{
		DB:  conn,
		Ctx: ctx,
	}, closer
}

func (s *store) Migrate(schemaFile string) error {
	schema, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Fatalln("can't read SQL schema file.", err)
	}
	query := string(schema)
	if _, err := s.DB.Exec(s.Ctx, query); err == nil {
		log.Println("DB schema has been created")
	} else {
		log.Fatalln("failed creating DB schema:", err)
	}
	return nil
}

func (s store) Ping(ctx context.Context) error {
	return s.DB.Ping(ctx)
}

func (s store) Get(metricType string, metricName string) (models.Metrics, error) {
	return models.Metrics{}, nil
}

func (s store) GetAll() []models.Metrics {
	return nil
}

func (s *store) Update(m models.Metrics) error {
	return nil
}
