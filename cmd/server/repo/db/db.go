package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type store struct {
	DB  *pgxpool.Pool
	Ctx context.Context
}

func New(ctx context.Context, envCfg *config.Config) (*store, func()) {
	conn, err := pgxpool.Connect(ctx, envCfg.PgDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	closer := func() {
		conn.Close()
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

func (s *store) BatchUpdate(ctx context.Context, metrics []models.Metrics) error {
	return nil
}

func (s *store) Dump(ctx context.Context, metrics []models.Metrics) error {
	log.Println("Dumping to Postgres...")
	q := `INSERT INTO metrics (type, name, value, delta)
			  VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO UPDATE SET
	      value = excluded.value, delta = excluded.delta;`

	for _, m := range metrics {
		_, err := s.DB.Exec(context.Background(), q, m.MType, m.ID, m.Value, m.Delta)
		if err != nil {
			return fmt.Errorf("failed dumping metric `%#v`. %w", m, err)
		}
	}
	log.Println("Metrics dumped into PostgreSQL.")
	return nil
}

func (s *store) Restore() ([]models.Metrics, error) {
	metrics, err := s.getMetrics()
	if err != nil {
		return nil, err
	}

	log.Println("Metrics restored from PostgreSQL.")
	return metrics, nil
}

func (s *store) getMetrics() ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0, 10)

	rows, err := s.DB.Query(context.Background(), "select type, name, value, delta from metrics")
	if err != nil {
		return metrics, err
	}
	defer rows.Close()

	for rows.Next() {
		m := new(models.Metrics)
		err := rows.Scan(&m.MType, &m.ID, &m.Value, &m.Delta)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, *m)
	}

	return metrics, nil
}

func (s *store) BatchUpsert(metrics []models.Metrics) error {
	// TODO: implement. Just call Dump?
	return nil
}
