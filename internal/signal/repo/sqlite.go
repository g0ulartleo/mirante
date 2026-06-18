package repo

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/g0ulartleo/mirante/internal/signal"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteSignalRepository struct {
	db *sql.DB
}

func NewSQLiteSignalRepository() (signal.SignalRepository, error) {
	db, err := sql.Open("sqlite3", "sqlite.db")
	if err != nil {
		return nil, err
	}
	repo := &SQLiteSignalRepository{db: db}
	if err := repo.Init(); err != nil {
		return nil, err
	}

	return &SQLiteSignalRepository{db: db}, nil
}

func (r *SQLiteSignalRepository) Save(signal signal.Signal) error {
	detailsJSON, err := json.Marshal(signal.Details)
	if err != nil {
		return err
	}
	query := `INSERT INTO signals (alarm_id, status, message, details_json, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err = r.db.Exec(query, signal.AlarmID, signal.Status, signal.Message, string(detailsJSON), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLiteSignalRepository) GetAlarmLatestSignals(alarmID string, limit int) ([]signal.Signal, error) {
	query := `
		SELECT alarm_id, status, message, details_json, created_at
		FROM signals WHERE alarm_id = ? ORDER BY created_at DESC LIMIT ?`
	rows, err := r.db.Query(query, alarmID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	signals := make([]signal.Signal, 0)
	for rows.Next() {
		var s signal.Signal
		var detailsJSON sql.NullString
		err := rows.Scan(&s.AlarmID, &s.Status, &s.Message, &detailsJSON, &s.Timestamp)
		if err != nil {
			return nil, err
		}
		if err := scanDetails(detailsJSON, &s); err != nil {
			return nil, err
		}
		signals = append(signals, s)
	}
	return signals, nil
}

func (r *SQLiteSignalRepository) GetAlarmHealth(alarmID string) (signal.Status, error) {
	query := `
		SELECT status
		FROM signals
		WHERE alarm_id = ?
		ORDER BY created_at DESC
		LIMIT 1`
	rows, err := r.db.Query(query, alarmID)
	if err != nil {
		return signal.StatusUnknown, err
	}
	defer rows.Close()
	if rows.Next() {
		var status signal.Status
		err := rows.Scan(&status)
		if err != nil {
			return signal.StatusUnknown, err
		}
		return status, nil
	}
	return signal.StatusUnknown, nil
}

func (r *SQLiteSignalRepository) CountUnhealthySince(alarmID string, since time.Time) (int, error) {
	query := `
        SELECT COUNT(*)
        FROM signals
        WHERE alarm_id = ? AND status = 'unhealthy' AND created_at >= ?`
	var count int
	err := r.db.QueryRow(query, alarmID, since).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SQLiteSignalRepository) Init() error {
	query := `CREATE TABLE IF NOT EXISTS signals (
		alarm_id VARCHAR(255) NOT NULL,
		status VARCHAR(255) NOT NULL,
		message VARCHAR(255) NOT NULL,
		details_json TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`ALTER TABLE signals ADD COLUMN details_json TEXT`)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return err
	}
	return nil
}

func (r *SQLiteSignalRepository) CleanOldSignals() error {
	query := `DELETE FROM signals WHERE created_at < NOW() - INTERVAL 14 DAY`
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLiteSignalRepository) Close() error {
	return r.db.Close()
}
