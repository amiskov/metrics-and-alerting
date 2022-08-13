package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
)

type Storage struct {
	DB *pgx.Conn
}

func Migrate(conn *pgx.Conn, schemaFile string) error {
	schema, err := os.ReadFile(schemaFile)
	if err != nil {
		log.Fatalln("can't read SQL schema file.", err)
	}
	query := string(schema)
	if _, err := conn.Exec(context.Background(), query); err == nil {
		log.Println("DB schema has been created")
	} else {
		log.Fatalln("failed creating DB schema:", err)
	}
	return nil
}

func (s Storage) Ping(ctx context.Context) error {
	return s.DB.Ping(ctx)
}
