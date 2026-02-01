package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Device struct {
	ID        int64
	IP        string
	MAC       string
	Hostname  string
	Vendor    string
	IsOnline  bool
	FirstSeen time.Time
	LastSeen  time.Time
}

type Stats struct {
	TotalDevices   int
	OnlineDevices  int
	OfflineDevices int
}

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT NOT NULL UNIQUE,
		mac_address TEXT,
		hostname TEXT,
		vendor TEXT,
		is_online BOOLEAN DEFAULT 1,
		first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_devices_ip ON devices(ip_address);
	CREATE INDEX IF NOT EXISTS idx_devices_mac ON devices(mac_address);
	CREATE INDEX IF NOT EXISTS idx_devices_online ON devices(is_online);
	`

	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) SaveDevice(ip, mac, hostname, vendor string) error {
	query := `
	INSERT INTO devices (ip_address, mac_address, hostname, vendor, is_online, last_seen_at)
	VALUES (?, ?, ?, ?, 1, ?)
	ON CONFLICT(ip_address) DO UPDATE SET
		mac_address = CASE
			WHEN excluded.mac_address != '' THEN excluded.mac_address
			ELSE mac_address
		END,
		hostname = CASE
			WHEN excluded.hostname != '' THEN excluded.hostname
			ELSE hostname
		END,
		vendor = CASE
			WHEN excluded.vendor != '' THEN excluded.vendor
			ELSE vendor
		END,
		is_online = 1,
		last_seen_at = excluded.last_seen_at
	`

	_, err := db.conn.Exec(query, ip, mac, hostname, vendor, time.Now())
	return err
}

func (db *DB) GetAllDevices() ([]Device, error) {
	query := `
	SELECT id, ip_address, mac_address, hostname, vendor, is_online, first_seen_at, last_seen_at
	FROM devices
	ORDER BY is_online DESC, ip_address
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		var mac, hostname, vendor sql.NullString

		err := rows.Scan(&d.ID, &d.IP, &mac, &hostname, &vendor, &d.IsOnline, &d.FirstSeen, &d.LastSeen)
		if err != nil {
			return nil, err
		}

		d.MAC = mac.String
		d.Hostname = hostname.String
		d.Vendor = vendor.String

		devices = append(devices, d)
	}

	return devices, rows.Err()
}


func (db *DB) MarkStaleDevicesOffline(minutes int) error {
	query := `
	UPDATE devices
	SET is_online = 0
	WHERE datetime(last_seen_at) < datetime('now', '-' || ? || ' minutes')
	AND is_online = 1
	`

	_, err := db.conn.Exec(query, minutes)
	return err
}

func (db *DB) GetStats() (Stats, error) {
	var stats Stats

	err := db.conn.QueryRow("SELECT COUNT(*) FROM devices").Scan(&stats.TotalDevices)
	if err != nil {
		return stats, err
	}

	err = db.conn.QueryRow("SELECT COUNT(*) FROM devices WHERE is_online = 1").Scan(&stats.OnlineDevices)
	if err != nil {
		return stats, err
	}

	stats.OfflineDevices = stats.TotalDevices - stats.OnlineDevices

	return stats, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
