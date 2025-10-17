package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type User struct {
	ID      int     `db:"id"`
	Name    string  `db:"name"`
	Email   string  `db:"email"`
	Balance float64 `db:"balance"`
}

func mustConnect() *sqlx.DB {
	dsn := "postgres://user:password@localhost:5430/mydatabase?sslmode=disable"

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	return db
}

func InsertUser(db *sqlx.DB, user User) error {
	const q = `
		INSERT INTO users (name, email, balance)
		VALUES (:name, :email, :balance)
	`
	_, err := db.NamedExec(q, user)
	return err
}

func GetAllUsers(db *sqlx.DB) ([]User, error) {
	const q = `SELECT id, name, email, balance FROM users ORDER BY id`
	var users []User
	if err := db.Select(&users, q); err != nil {
		return nil, err
	}
	return users, nil
}

func GetUserByID(db *sqlx.DB, id int) (User, error) {
	const q = `SELECT id, name, email, balance FROM users WHERE id = $1`
	var u User
	err := db.Get(&u, q, id)
	return u, err
}

func TransferBalance(db *sqlx.DB, fromID int, toID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be > 0")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var from User
	if err = tx.GetContext(ctx, &from, `SELECT id, name, email, balance FROM users WHERE id=$1 FOR UPDATE`, fromID); err != nil {
		return fmt.Errorf("get sender: %w", err)
	}
	var to User
	if err = tx.GetContext(ctx, &to, `SELECT id, name, email, balance FROM users WHERE id=$1 FOR UPDATE`, toID); err != nil {
		return fmt.Errorf("get receiver: %w", err)
	}

	if from.Balance < amount {
		return fmt.Errorf("insufficient funds: have %.2f, need %.2f", from.Balance, amount)
	}

	if _, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance - $1 WHERE id=$2`, amount, fromID); err != nil {
		return fmt.Errorf("debit sender: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance + $1 WHERE id=$2`, amount, toID); err != nil {
		return fmt.Errorf("credit receiver: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func main() {
	db := mustConnect()
	defer db.Close()
	fmt.Println("Connected to Postgres via sqlx ✔")

	seed := []User{
		{Name: "Айман", Email: "aizhan@example.com", Balance: 100},
		{Name: "Ермек", Email: "ermek@example.com", Balance: 50},
		{Name: "Бекзат", Email: "bekzat@example.com", Balance: 25},
	}

	for _, u := range seed {
		if err := InsertUser(db, u); err != nil {
			log.Fatalf("InsertUser(%s): %v", u.Email, err)
		}
	}
	fmt.Println("Inserted 3 users")

	users, err := GetAllUsers(db)
	if err != nil {
		log.Fatalf("GetAllUsers: %v", err)
	}
	fmt.Println("All users:")
	for _, u := range users {
		fmt.Printf("  #%d %s | %s | balance=%.2f\n", u.ID, u.Name, u.Email, u.Balance)
	}

	u1, err := GetUserByID(db, users[0].ID)
	if err != nil {
		log.Fatalf("GetUserByID: %v", err)
	}
	fmt.Printf("GetUserByID(%d): %+v\n", u1.ID, u1)

	if err := TransferBalance(db, users[0].ID, users[1].ID, 15); err != nil {
		log.Fatalf("TransferBalance: %v", err)
	}
	fmt.Println("Transfer 15: OK")

	users, _ = GetAllUsers(db)
	fmt.Println("After transfer:")
	for _, u := range users {
		fmt.Printf("  #%d %s | balance=%.2f\n", u.ID, u.Name, u.Balance)
	}
}
