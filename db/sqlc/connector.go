package db

import "github.com/jackc/pgx/v5/pgxpool"

type DBConnector struct {
    Querier
    db      *pgxpool.Pool
}

// NewConnector function to initialize the Connector
func NewConnector(db *pgxpool.Pool) *DBConnector {
    return &DBConnector{
        db:      db,
        Querier: New(db),
    }
}
