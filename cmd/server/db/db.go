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

func (s *store) Dump(inmemDB models.InmemDB) error {
	q := `INSERT INTO metrics (type, name, value, delta, hash)
			  VALUES ($1, $2, $3, $4, $5) ON CONFLICT (name) DO UPDATE SET
	      value = excluded.value, delta = excluded.delta, hash = excluded.hash;`

	for _, m := range inmemDB.GetAll() {
		ct, err := s.DB.Exec(context.Background(), q, m.MType, m.ID, m.Value, m.Delta, m.Hash)
		if err != nil {
			log.Printf("failed dumping metric `%#v`. CommantTag: `%#v`, err: `%#v`.\n", m, string(ct), err)
			return err
		}
	}
	log.Println("Metrics dumped into PostgreSQL.")
	return nil
}

func (s *store) Restore(inmemDB models.InmemDB) error {
	metrics, err := s.getMetrics()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		err := inmemDB.Upsert(m)
		if err != nil {
			return err
		}
	}

	log.Println("Metrics restored from PostgreSQL.")
	return nil
}

func (s *store) getMetrics() ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0, 10)

	rows, err := s.DB.Query(context.Background(), "select type, name, value, delta, hash from metrics")
	if err != nil {
		return metrics, err
	}
	defer rows.Close()

	for rows.Next() {
		m := new(models.Metrics)
		err := rows.Scan(&m.MType, &m.ID, &m.Value, &m.Delta, &m.Hash)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, *m)
	}

	return metrics, nil
}
