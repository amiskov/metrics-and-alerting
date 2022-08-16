package postgres

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

const insertMetricQuery = `INSERT INTO metrics (type, name, value, delta)
	VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO UPDATE SET
	value = excluded.value, delta = excluded.delta;`

type db struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

func New(ctx context.Context, envCfg *config.Config) (*db, func()) {
	conn, err := pgxpool.Connect(ctx, envCfg.PgDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	s := &db{
		pool: conn,
		ctx:  ctx,
	}
	return s, func() { conn.Close() }
}

// Creates the `metrics` table if not exists.
func (d *db) Migrate() {
	// Check if table exists
	_, err := d.pool.Exec(context.Background(), "select id, type, name, value, delta from metrics where id = 1")
	if err == nil {
		return
	}

	schema, err := os.ReadFile("sql/schema.sql")
	if err != nil {
		log.Fatalln("can't read SQL schema file.", err)
	}

	_, exErr := d.pool.Exec(d.ctx, string(schema))
	if exErr != nil {
		log.Fatalln("failed creating DB schema:", exErr)
	}

	log.Println("DB schema has been created")
}

func (d *db) Get(metricType string, metricName string) (models.Metrics, error) {
	m := models.Metrics{}

	q := "select type, name, value, delta from metrics where type = $1 and name = $2"
	row := d.pool.QueryRow(d.ctx, q, metricType, metricName)
	err := row.Scan(&m.MType, &m.ID, &m.Value, &m.Delta)
	if err != nil {
		return m, err
	}

	return m, nil
}

func (d *db) GetAll() ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0, 10)

	rows, err := d.pool.Query(context.Background(), "select type, name, value, delta from metrics")
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

func (d db) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

func (d *db) Update(ctx context.Context, m models.Metrics) error {
	_, err := d.pool.Exec(context.Background(), insertMetricQuery, m.MType, m.ID, m.Value, m.Delta)
	if err != nil {
		return fmt.Errorf("failed inserting metric `%#v`. %w", m, err)
	}
	return nil
}

func (d *db) BulkUpdate(metrics []models.Metrics) error {
	tx, err := d.pool.Begin(d.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(d.ctx)

	preparedStatementName := "bulkUpdate"
	if _, prepErr := tx.Prepare(d.ctx, preparedStatementName, insertMetricQuery); prepErr != nil {
		return fmt.Errorf("pg: failed preparing transaction statement: %w", prepErr)
	}

	for _, m := range metrics {
		if _, err = tx.Exec(d.ctx, preparedStatementName, m.MType, m.ID, m.Value, m.Delta); err != nil {
			return fmt.Errorf("pg: failed executing transaction: %w", err)
		}
	}
	return tx.Commit(d.ctx)
}
