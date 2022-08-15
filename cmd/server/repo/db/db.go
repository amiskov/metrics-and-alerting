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
	db         *pgxpool.Pool
	ctx        context.Context
	hashingKey []byte
}

func New(ctx context.Context, envCfg *config.Config) (*store, func()) {
	conn, err := pgxpool.Connect(ctx, envCfg.PgDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	s := &store{
		db:         conn,
		ctx:        ctx,
		hashingKey: []byte(envCfg.HashingKey),
	}
	return s, func() { conn.Close() }
}

// Creates the `metrics` table if not exists.
func (s *store) Migrate() {
	// Check if table exists
	_, err := s.db.Exec(context.Background(), "select id, type, name, value, delta from metrics where id = 1")
	if err == nil {
		return
	}

	schema, err := os.ReadFile("sql/schema.sql")
	if err != nil {
		log.Fatalln("can't read SQL schema file.", err)
	}

	_, exErr := s.db.Exec(s.ctx, string(schema))
	if exErr != nil {
		log.Fatalln("failed creating DB schema:", exErr)
	}

	log.Println("DB schema has been created")
}

func (s store) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *store) Upsert(ctx context.Context, m models.Metrics) error {
	q := `INSERT INTO metrics (type, name, value, delta)
			  VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO UPDATE SET
	      value = excluded.value, delta = excluded.delta;`

	_, err := s.db.Exec(context.Background(), q, m.MType, m.ID, m.Value, m.Delta)
	if err != nil {
		return fmt.Errorf("failed upserting metric `%#v`. %w", m, err)
	}
	return nil
}

func (s *store) Dump(ctx context.Context, metrics []models.Metrics) error {
	// TODO: use batch inserting
	for _, m := range metrics {
		err := s.Upsert(ctx, m)
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

func (s *store) Get(metricType string, metricName string) (models.Metrics, error) {
	m := models.Metrics{}
	row := s.db.QueryRow(s.ctx, "select type, name, value, delta from metrics where type = $1 and name = $2",
		metricType, metricName)
	err := row.Scan(&m.MType, &m.ID, &m.Value, &m.Delta)
	if err != nil {
		return m, err
	}

	if len(s.hashingKey) > 0 {
		hash, err := m.GetHash(s.hashingKey)
		if err != nil {
			log.Println("can't actualize hash", err)
			return m, err
		}
		m.Hash = hash
	}

	return m, nil
}

func (s *store) GetAll() []models.Metrics {
	metrics, err := s.getMetrics()
	if err != nil {
		log.Println("failed get metrics from Postgres", err)
		return nil
	}

	if len(s.hashingKey) > 0 {
		for k, m := range metrics {
			hash, err := m.GetHash(s.hashingKey)
			if err != nil {
				log.Println("can't actualize hash", err)
				break
			}
			m.Hash = hash
			metrics[k] = m
		}
	}

	return metrics
}

func (s *store) getMetrics() ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0, 10)

	rows, err := s.db.Query(context.Background(), "select type, name, value, delta from metrics")
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
