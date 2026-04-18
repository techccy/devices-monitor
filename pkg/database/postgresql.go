package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgreSQLStorage struct {
	db *sql.DB
}

func NewPostgreSQLStorage(dsn string) (*PostgreSQLStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgreSQLStorage{db: db}, nil
}

func (p *PostgreSQLStorage) Close() error {
	return p.db.Close()
}

func (p *PostgreSQLStorage) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(255) PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS devices (
		id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		identifier VARCHAR(255) NOT NULL,
		device_key VARCHAR(255) NOT NULL,
		online BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS snapshots (
		device_id VARCHAR(255) PRIMARY KEY,
		timestamp BIGINT NOT NULL,
		cpu_usage FLOAT,
		memory_usage FLOAT,
		disk_usage FLOAT,
		disk_remaining BIGINT,
		network_status TEXT,
		FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS json_documents (
		id SERIAL PRIMARY KEY,
		collection VARCHAR(255) NOT NULL,
		data JSONB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_devices_user_id ON devices(user_id);
	CREATE INDEX IF NOT EXISTS idx_snapshots_device_id ON snapshots(device_id);
	CREATE INDEX IF NOT EXISTS idx_json_documents_collection ON json_documents(collection);
	`

	_, err := p.db.Exec(schema)
	return err
}

func (p *PostgreSQLStorage) CreateUser(email, password string) (string, error) {
	id := generateID()
	query := `INSERT INTO users (id, email, password) VALUES ($1, $2, $3)`
	_, err := p.db.Exec(query, id, email, password)
	return id, err
}

func (p *PostgreSQLStorage) GetUserByEmail(email string) (map[string]interface{}, error) {
	query := `SELECT id, email, password FROM users WHERE email = $1`
	row := p.db.QueryRow(query, email)

	var id, emailResult, password string
	err := row.Scan(&id, &emailResult, &password)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":       id,
		"email":    emailResult,
		"password": password,
	}, nil
}

func (p *PostgreSQLStorage) GetUser(id string) (map[string]interface{}, error) {
	query := `SELECT id, email, password FROM users WHERE id = $1`
	row := p.db.QueryRow(query, id)

	var email, password string
	err := row.Scan(&id, &email, &password)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":       id,
		"email":    email,
		"password": password,
	}, nil
}

func (p *PostgreSQLStorage) CreateDevice(userID, name, identifier string) (map[string]interface{}, error) {
	id := generateID()
	deviceKey := generateID()
	query := `INSERT INTO devices (id, user_id, name, identifier, device_key) VALUES ($1, $2, $3, $4, $5)`
	_, err := p.db.Exec(query, id, userID, name, identifier, deviceKey)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":         id,
		"user_id":    userID,
		"name":       name,
		"identifier": identifier,
		"device_key": deviceKey,
	}, nil
}

func (p *PostgreSQLStorage) GetDevice(id string) (map[string]interface{}, error) {
	query := `SELECT id, user_id, name, identifier, device_key, online, created_at, updated_at FROM devices WHERE id = $1`
	row := p.db.QueryRow(query, id)

	var userID, name, identifier, deviceKey string
	var online bool
	var createdAt, updatedAt time.Time
	err := row.Scan(&id, &userID, &name, &identifier, &deviceKey, &online, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":         id,
		"user_id":    userID,
		"name":       name,
		"identifier": identifier,
		"device_key": deviceKey,
		"online":     online,
		"created_at": createdAt.Unix(),
		"updated_at": updatedAt.Unix(),
	}, nil
}

func (p *PostgreSQLStorage) GetDevicesByUser(userID string) ([]map[string]interface{}, error) {
	query := `SELECT id, user_id, name, identifier, device_key, online, created_at, updated_at FROM devices WHERE user_id = $1`
	rows, err := p.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []map[string]interface{}
	for rows.Next() {
		var id, userIDResult, name, identifier, deviceKey string
		var online bool
		var createdAt, updatedAt time.Time
		err := rows.Scan(&id, &userIDResult, &name, &identifier, &deviceKey, &online, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		devices = append(devices, map[string]interface{}{
			"id":         id,
			"user_id":    userIDResult,
			"name":       name,
			"identifier": identifier,
			"device_key": deviceKey,
			"online":     online,
			"created_at": createdAt.Unix(),
			"updated_at": updatedAt.Unix(),
		})
	}

	return devices, nil
}

func (p *PostgreSQLStorage) UpdateDeviceStatus(deviceID string, online bool) error {
	query := `UPDATE devices SET online = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := p.db.Exec(query, online, deviceID)
	return err
}

func (p *PostgreSQLStorage) SaveSnapshot(deviceID string, snapshotData interface{}) error {
	snapshotMap, ok := snapshotData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid snapshot data type")
	}

	query := `
	INSERT INTO snapshots (device_id, timestamp, cpu_usage, memory_usage, disk_usage, disk_remaining, network_status)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (device_id) DO UPDATE SET
		timestamp = EXCLUDED.timestamp,
		cpu_usage = EXCLUDED.cpu_usage,
		memory_usage = EXCLUDED.memory_usage,
		disk_usage = EXCLUDED.disk_usage,
		disk_remaining = EXCLUDED.disk_remaining,
		network_status = EXCLUDED.network_status
	`

	var cpuUsage, memoryUsage, diskUsage sql.NullFloat64
	var diskRemaining sql.NullInt64
	var networkStatus sql.NullString

	if val, ok := snapshotMap["cpu_usage"].(float64); ok {
		cpuUsage = sql.NullFloat64{Float64: val, Valid: true}
	}
	if val, ok := snapshotMap["memory_usage"].(float64); ok {
		memoryUsage = sql.NullFloat64{Float64: val, Valid: true}
	}
	if val, ok := snapshotMap["disk_usage"].(float64); ok {
		diskUsage = sql.NullFloat64{Float64: val, Valid: true}
	}
	if val, ok := snapshotMap["disk_remaining"].(int64); ok {
		diskRemaining = sql.NullInt64{Int64: val, Valid: true}
	}
	if val, ok := snapshotMap["network_status"].(string); ok {
		networkStatus = sql.NullString{String: val, Valid: true}
	}

	timestamp := time.Now().Unix()
	_, err := p.db.Exec(query, deviceID, timestamp, cpuUsage, memoryUsage, diskUsage, diskRemaining, networkStatus)
	return err
}

func (p *PostgreSQLStorage) GetSnapshot(deviceID string) (map[string]interface{}, error) {
	query := `SELECT device_id, timestamp, cpu_usage, memory_usage, disk_usage, disk_remaining, network_status FROM snapshots WHERE device_id = $1`
	row := p.db.QueryRow(query, deviceID)

	var deviceIDResult string
	var timestamp int64
	var cpuUsage, memoryUsage, diskUsage sql.NullFloat64
	var diskRemaining sql.NullInt64
	var networkStatus sql.NullString

	err := row.Scan(&deviceIDResult, &timestamp, &cpuUsage, &memoryUsage, &diskUsage, &diskRemaining, &networkStatus)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"device_id": deviceIDResult,
		"timestamp": timestamp,
	}

	if cpuUsage.Valid {
		result["cpu_usage"] = cpuUsage.Float64
	}
	if memoryUsage.Valid {
		result["memory_usage"] = memoryUsage.Float64
	}
	if diskUsage.Valid {
		result["disk_usage"] = diskUsage.Float64
	}
	if diskRemaining.Valid {
		result["disk_remaining"] = diskRemaining.Int64
	}
	if networkStatus.Valid {
		result["network_status"] = networkStatus.String
	}

	return result, nil
}

func (p *PostgreSQLStorage) CreateJSONDocument(collection string, document map[string]interface{}) error {
	data, err := json.Marshal(document)
	if err != nil {
		return err
	}

	query := `INSERT INTO json_documents (collection, data) VALUES ($1, $2)`
	_, err = p.db.Exec(query, collection, data)
	return err
}

func (p *PostgreSQLStorage) FindJSONDocuments(collection string, filter map[string]interface{}) ([]map[string]interface{}, error) {
	query := `SELECT data FROM json_documents WHERE collection = $1`
	rows, err := p.db.Query(query, collection)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []map[string]interface{}
	for rows.Next() {
		var data []byte
		err := rows.Scan(&data)
		if err != nil {
			return nil, err
		}

		var doc map[string]interface{}
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, err
		}

		if matchesFilter(doc, filter) {
			documents = append(documents, doc)
		}
	}

	return documents, nil
}

func matchesFilter(doc, filter map[string]interface{}) bool {
	for key, value := range filter {
		if docValue, exists := doc[key]; !exists || docValue != value {
			return false
		}
	}
	return true
}

func (p *PostgreSQLStorage) CreateJSONIndex(collection, field string) error {
	query := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_%s ON json_documents ((data->'%s'))`, collection, field, field)
	_, err := p.db.Exec(query)
	return err
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
