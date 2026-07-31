package oui

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

const ieeeCSVURL = "https://standards-oui.ieee.org/oui/oui.csv"

type DB struct {
	mu      sync.RWMutex
	entries map[string]string
	custom  map[string]string
}

func New() *DB {
	return &DB{
		entries: make(map[string]string),
		custom:  make(map[string]string),
	}
}

func (d *DB) Lookup(mac string) string {
	prefix := normalize(mac)
	if len(prefix) < 6 {
		return ""
	}
	ouiKey := prefix[:6]

	d.mu.RLock()
	defer d.mu.RUnlock()

	if v, ok := d.custom[ouiKey]; ok {
		return v
	}
	if v, ok := d.entries[ouiKey]; ok {
		return v
	}
	return ""
}

func (d *DB) Update() error {
	resp, err := http.Get(ieeeCSVURL)
	if err != nil {
		return fmt.Errorf("download oui csv: %w", err)
	}
	defer resp.Body.Close()

	entries, err := parseCSV(resp.Body)
	if err != nil {
		return err
	}

	d.mu.Lock()
	d.entries = entries
	d.mu.Unlock()
	return nil
}

func (d *DB) Entries() map[string]string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	m := make(map[string]string, len(d.entries))
	for k, v := range d.entries {
		m[k] = v
	}
	return m
}

func (d *DB) Custom() map[string]string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	m := make(map[string]string, len(d.custom))
	for k, v := range d.custom {
		m[k] = v
	}
	return m
}

func (d *DB) SetCustom(vendors map[string]string) {
	d.mu.Lock()
	d.custom = make(map[string]string, len(vendors))
	for k, v := range vendors {
		d.custom[strings.ToUpper(strings.ReplaceAll(k, "-", ""))] = v
	}
	d.mu.Unlock()
}

func parseCSV(r io.Reader) (map[string]string, error) {
	entries := make(map[string]string)
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1

	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(record) < 3 {
			continue
		}
		registry := strings.TrimSpace(record[0])
		assignment := strings.TrimSpace(record[1])
		org := strings.TrimSpace(record[2])

		if org == "" || !strings.Contains(registry, "MA-L") {
			continue
		}

		prefix := normalizeAssignment(assignment)
		if prefix == "" {
			continue
		}
		if _, exists := entries[prefix]; !exists {
			entries[prefix] = org
		}
	}
	return entries, nil
}

func normalizeAssignment(a string) string {
	a = strings.ReplaceAll(a, "-", "")
	a = strings.ReplaceAll(a, ":", "")
	a = strings.ToUpper(a)
	if len(a) < 6 {
		return ""
	}
	return a[:6]
}

func normalize(mac string) string {
	var b strings.Builder
	for _, c := range mac {
		if c != ':' && c != '-' && c != ' ' && c != '.' {
			b.WriteRune(c)
		}
	}
	return strings.ToUpper(b.String())
}
