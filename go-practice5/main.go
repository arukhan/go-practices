package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Product struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Price    int    `json:"price"`
}

var db *pgxpool.Pool

func main() {
	var err error

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/products_db?sslmode=disable"
	}

	db, err = pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer db.Close()

	if err = db.Ping(context.Background()); err != nil {
		log.Fatal("Cannot ping database:", err)
	}
	log.Println("Connected to database")

	http.HandleFunc("/products", getProductsHandler)

	addr := ":8080"
	log.Println("listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func getProductsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	sb := strings.Builder{}
	sb.WriteString(`SELECT p.id, p.name, c.name AS category, p.price
FROM products p
JOIN categories c ON c.id = p.category_id`)

	clauses := []string{}
	args := []interface{}{}
	i := 1

	q := r.URL.Query()

	if v := q.Get("category"); v != "" {
		clauses = append(clauses, fmt.Sprintf("c.name = $%d", i))
		args = append(args, v)
		i++
	}
	if v := q.Get("min_price"); v != "" {
		minp, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, "min_price must be integer", http.StatusBadRequest)
			return
		}
		clauses = append(clauses, fmt.Sprintf("p.price >= $%d", i))
		args = append(args, minp)
		i++
	}
	if v := q.Get("max_price"); v != "" {
		maxp, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, "max_price must be integer", http.StatusBadRequest)
			return
		}
		clauses = append(clauses, fmt.Sprintf("p.price <= $%d", i))
		args = append(args, maxp)
		i++
	}

	if len(clauses) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(clauses, " AND "))
	}

	switch q.Get("sort") {
	case "price_asc":
		sb.WriteString(" ORDER BY p.price ASC")
	case "price_desc":
		sb.WriteString(" ORDER BY p.price DESC")
	default:
		sb.WriteString(" ORDER BY p.id ASC")
	}

	limit := 50
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		} else {
			http.Error(w, "limit must be positive integer", http.StatusBadRequest)
			return
		}
	}
	offset := 0
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		} else {
			http.Error(w, "offset must be non-negative integer", http.StatusBadRequest)
			return
		}
	}

	sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1))
	args = append(args, limit, offset)

	sqlStr := sb.String()

	rows, err := db.Query(ctx, sqlStr, args...)
	if err != nil {
		log.Printf("query error: %v", err)
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := make([]Product, 0, limit)
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			return
		}
		results = append(results, p)
	}
	if rows.Err() != nil {
		http.Error(w, "rows error", http.StatusInternalServerError)
		return
	}

	dur := time.Since(start)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Query-Time", dur.String())

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(results); err != nil {
		http.Error(w, "encode failed", http.StatusInternalServerError)
		return
	}

	log.Printf("OK %s args=%v rows=%d took=%s", r.URL.RawQuery, args, len(results), dur)
}
