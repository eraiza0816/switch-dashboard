package history

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

type HistoryPoint struct {
	TS float64 `json:"ts"`
	TX int64   `json:"tx"`
	RX int64   `json:"rx"`
}

type Store struct {
	db *sql.DB
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping duckdb: %w", err)
	}
	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Retain(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	_, err := s.db.Exec(`DELETE FROM bandwidth_samples WHERE ts < $1`, cutoff)
	return err
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS bandwidth_samples (
			ts        TIMESTAMP NOT NULL,
			switch_ip VARCHAR   NOT NULL,
			port      VARCHAR   NOT NULL,
			cum_tx    BIGINT    NOT NULL,
			cum_rx    BIGINT    NOT NULL,
			speed_tx  BIGINT    NOT NULL,
			speed_rx  BIGINT    NOT NULL
		)
	`)
	return err
}

func (s *Store) WriteSample(ip, port string, ts time.Time, cumTX, cumRX, speedTX, speedRX int64) error {
	_, err := s.db.Exec(
		`INSERT INTO bandwidth_samples (ts, switch_ip, port, cum_tx, cum_rx, speed_tx, speed_rx) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		ts.UTC(), ip, port, cumTX, cumRX, speedTX, speedRX,
	)
	return err
}

func (s *Store) QueryHistory(ip, port, rng string) ([]HistoryPoint, error) {
	var since time.Duration
	switch rng {
	case "live":
		since = 2 * time.Minute
	case "1h":
		since = 1 * time.Hour
	case "24h":
		since = 24 * time.Hour
	default:
		since = 1 * time.Hour
	}

	cutoff := time.Now().Add(-since)

	if rng == "24h" {
		return s.queryAggregated(ip, port, cutoff)
	}
	return s.queryRaw(ip, port, cutoff)
}

func (s *Store) queryRaw(ip, port string, cutoff time.Time) ([]HistoryPoint, error) {
	rows, err := s.db.Query(
		`SELECT ts, speed_tx, speed_rx FROM bandwidth_samples WHERE switch_ip = $1 AND port = $2 AND ts >= $3 ORDER BY ts`,
		ip, port, cutoff,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var points []HistoryPoint
	for rows.Next() {
		var ts time.Time
		var tx, rx int64
		if err := rows.Scan(&ts, &tx, &rx); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		points = append(points, HistoryPoint{
			TS: float64(ts.UnixNano()) / 1e9,
			TX: tx,
			RX: rx,
		})
	}
	if points == nil {
		points = []HistoryPoint{}
	}
	return points, nil
}

func (s *Store) queryAggregated(ip, port string, cutoff time.Time) ([]HistoryPoint, error) {
	rows, err := s.db.Query(
		`SELECT date_trunc('hour', ts) as hour, CAST(avg(speed_tx) AS BIGINT) as avg_tx, CAST(avg(speed_rx) AS BIGINT) as avg_rx FROM bandwidth_samples WHERE switch_ip = $1 AND port = $2 AND ts >= $3 GROUP BY date_trunc('hour', ts) ORDER BY hour`,
		ip, port, cutoff,
	)
	if err != nil {
		return nil, fmt.Errorf("query aggregated: %w", err)
	}
	defer rows.Close()

	var points []HistoryPoint
	for rows.Next() {
		var ts time.Time
		var tx, rx int64
		if err := rows.Scan(&ts, &tx, &rx); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		points = append(points, HistoryPoint{
			TS: float64(ts.UnixNano()) / 1e9,
			TX: tx,
			RX: rx,
		})
	}
	if points == nil {
		points = []HistoryPoint{}
	}
	return points, nil
}
