package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var (
	db        *sql.DB
	redisURL  string
	mongoURL  string
	appName   string
)

type Item struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	appName = getEnv("APP_NAME", "go-postgres-service")
	redisURL = getEnv("REDIS_SERVICE_URL", "http://python-redis-service:8080")
	mongoURL = getEnv("MONGO_SERVICE_URL", "http://java-mongo-service:8080")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("POSTGRES_HOST", "postgres"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_USER", "appuser"),
		getEnv("POSTGRES_PASSWORD", "apppass"),
		getEnv("POSTGRES_DB", "appdb"),
	)

	var err error
	for i := 0; i < 30; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("DB connect attempt %d: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err = db.Ping(); err != nil {
			log.Printf("DB ping attempt %d: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		log.Fatalf("Could not connect to Postgres: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("Create table: %v", err)
	}

	http.HandleFunc("/health", health)
	http.HandleFunc("/items", itemsHandler)
	http.HandleFunc("/demo", demoHandler)

	port := getEnv("PORT", "8080")
	log.Printf("%s listening on :%s", appName, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func health(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": appName})
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query("SELECT id, name, created_at FROM items ORDER BY id DESC LIMIT 20")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var items []Item
		for rows.Next() {
			var it Item
			if err := rows.Scan(&it.ID, &it.Name, &it.CreatedAt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, it)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
		return
	case http.MethodPost:
		var body struct{ Name string `json:"name"` }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		var id int
		err := db.QueryRow("INSERT INTO items (name) VALUES ($1) RETURNING id", body.Name).Scan(&id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "name": body.Name})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func demoHandler(w http.ResponseWriter, r *http.Request) {
	// Demonstrates communication: Go -> Redis cache service -> Mongo document service
	result := map[string]interface{}{
		"service": appName,
		"postgres": "connected",
		"redis_service": nil,
		"mongo_service": nil,
	}

	// Call Python Redis service
	if resp, err := http.Get(redisURL + "/health"); err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		result["redis_service"] = string(body)
	} else {
		result["redis_service"] = "error: " + err.Error()
	}

	// Call Java Mongo service
	if resp, err := http.Get(mongoURL + "/health"); err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		result["mongo_service"] = string(body)
	} else {
		result["mongo_service"] = "error: " + err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
