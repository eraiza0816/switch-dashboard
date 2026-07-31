package history

import (
	"os"
	"testing"
	"time"
)

func TestDuckDBStore_WriteAndQuery(t *testing.T) {
	path := t.TempDir() + "/test.duckdb"
	defer os.Remove(path)

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	now := time.Now()
	// Write 5 samples for port 1 at 10s intervals
	for i := 0; i < 5; i++ {
		ts := now.Add(time.Duration(i*10) * time.Second)
		err := store.WriteSample("192.168.1.1", "1", ts, 1000+int64(i*100), 2000+int64(i*100), 800+int64(i*10), 1600+int64(i*10))
		if err != nil {
			t.Fatalf("WriteSample: %v", err)
		}
	}

	// Query live range
	points, err := store.QueryHistory("192.168.1.1", "1", "live")
	if err != nil {
		t.Fatalf("QueryHistory(live): %v", err)
	}
	if len(points) != 5 {
		t.Fatalf("live: expected 5 points, got %d", len(points))
	}

	// Verify data order (ascending by time)
	for i := 1; i < len(points); i++ {
		if points[i].TS <= points[i-1].TS {
			t.Fatalf("points not in ascending time order: %f <= %f", points[i].TS, points[i-1].TS)
		}
	}

	// Verify last point values
	last := points[len(points)-1]
	if last.TX != 840 || last.RX != 1640 {
		t.Fatalf("last point: expected TX=840, RX=1640, got TX=%d, RX=%d", last.TX, last.RX)
	}

	// Write samples for a different port and verify isolation
	err = store.WriteSample("192.168.1.1", "2", now, 500, 600, 400, 500)
	if err != nil {
		t.Fatalf("WriteSample port2: %v", err)
	}

	points2, err := store.QueryHistory("192.168.1.1", "2", "live")
	if err != nil {
		t.Fatalf("QueryHistory(port2): %v", err)
	}
	if len(points2) != 1 {
		t.Fatalf("port2: expected 1 point, got %d", len(points2))
	}

	// Query 1h range should include all points
	points1h, err := store.QueryHistory("192.168.1.1", "1", "1h")
	if err != nil {
		t.Fatalf("QueryHistory(1h): %v", err)
	}
	if len(points1h) != 5 {
		t.Fatalf("1h: expected 5 points, got %d", len(points1h))
	}
}

func TestDuckDBStore_EmptyQuery(t *testing.T) {
	path := t.TempDir() + "/empty.duckdb"
	defer os.Remove(path)

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	// Query for non-existent data should return empty slice
	points, err := store.QueryHistory("10.0.0.1", "5", "live")
	if err != nil {
		t.Fatalf("QueryHistory: %v", err)
	}
	if len(points) != 0 {
		t.Fatalf("expected 0 points, got %d", len(points))
	}
}

func TestDuckDBStore_24hQuery(t *testing.T) {
	path := t.TempDir() + "/24h.duckdb"
	defer os.Remove(path)

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	now := time.Now()
	// Write 48 samples (30 min apart = 24h)
	for i := 0; i < 48; i++ {
		ts := now.Add(-time.Duration(47-i) * 30 * time.Minute)
		err := store.WriteSample("10.0.0.1", "1", ts, 1000, 2000, 800, 1600)
		if err != nil {
			t.Fatalf("WriteSample: %v", err)
		}
	}

	points, err := store.QueryHistory("10.0.0.1", "1", "24h")
	if err != nil {
		t.Fatalf("QueryHistory(24h): %v", err)
	}
	if len(points) == 0 {
		t.Fatal("24h: expected at least 1 point")
	}

	// Verify hourly aggregation: 48 samples at 30min = 24 hours
	// Should produce ~24 hourly points
	if len(points) < 20 || len(points) > 26 {
		t.Fatalf("24h: expected ~24 hourly points, got %d", len(points))
	}

	// Verify we got data
	if len(points) > 0 {
		if points[0].TX < 0 || points[0].RX < 0 {
			t.Fatalf("point has negative value: TX=%d, RX=%d", points[0].TX, points[0].RX)
		}
	}
}

func TestDuckDBStore_DefaultRange(t *testing.T) {
	path := t.TempDir() + "/default.duckdb"
	defer os.Remove(path)

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	now := time.Now()
	store.WriteSample("10.0.0.1", "1", now, 1000, 2000, 800, 1600)

	// Empty range string should default to 1h
	points, err := store.QueryHistory("10.0.0.1", "1", "")
	if err != nil {
		t.Fatalf("QueryHistory(''): %v", err)
	}
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
}

func TestDuckDBStore_WriteAndQueryNilStore(t *testing.T) {
	path := t.TempDir() + "/nil.duckdb"
	defer os.Remove(path)

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	store.Close()

	// Writing to a closed store should fail
	err = store.WriteSample("10.0.0.1", "1", time.Now(), 1000, 2000, 800, 1600)
	if err == nil {
		t.Fatal("expected error writing to closed store")
	}
}

func TestDuckDBStore_Retention(t *testing.T) {
	path := t.TempDir() + "/retain.duckdb"
	defer os.Remove(path)

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	now := time.Now()
	// Write 3 old samples (25h ago) and 3 recent ones (1h ago)
	for i := 0; i < 3; i++ {
		ts := now.Add(-25 * time.Hour).Add(time.Duration(i) * time.Minute)
		store.WriteSample("10.0.0.1", "1", ts, 1000, 2000, 800, 1600)
	}
	for i := 0; i < 3; i++ {
		ts := now.Add(-1 * time.Hour).Add(time.Duration(i) * time.Minute)
		store.WriteSample("10.0.0.1", "1", ts, 1000, 2000, 800, 1600)
	}

	// Retain only last 24h
	err = store.Retain(24 * time.Hour)
	if err != nil {
		t.Fatalf("Retain: %v", err)
	}

	points, err := store.QueryHistory("10.0.0.1", "1", "24h")
	if err != nil {
		t.Fatalf("QueryHistory: %v", err)
	}
	if len(points) == 0 {
		t.Fatal("expected remaining points after retention")
	}

	// Query with no time limit to count all rows
	rows, err := store.db.Query("SELECT count(*) FROM bandwidth_samples")
	if err != nil {
		t.Fatalf("count query: %v", err)
	}
	defer rows.Close()
	var count int
	if rows.Next() {
		rows.Scan(&count)
	}
	if count != 3 {
		t.Fatalf("expected 3 rows after retention (oldest removed), got %d", count)
	}
}
