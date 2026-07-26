package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

func worker(
	wg *sync.WaitGroup,
	db *sql.DB,
	ctx context.Context,
	isolation sql.IsolationLevel,
	holdAmount int,
	mySignal chan struct{},
	otherSignal chan struct{},
) {

	defer wg.Done()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: isolation})
	if err != nil {
		log.Printf("hold %d: begin: %v", holdAmount, err)
		return
	}

	defer tx.Rollback()

	var available int

	if err := tx.QueryRow(`
		SELECT a.balance - COALESCE(SUM(h.amount),0)
		FROM accounts a
		LEFT JOIN pending_holds h ON h.account_id = a.id
		WHERE a.id = $1
		GROUP BY a.balance`, 1).Scan(&available); err != nil {
		log.Printf("hold%d: read: %v", holdAmount, err)
		return
	}
	log.Printf("hold%d: see: %v", holdAmount, available)

	mySignal <- struct{}{}
	<-otherSignal

	if holdAmount > available {
		log.Printf("hold%d: REJECTED (available=%v)", holdAmount, available)
		return
	}

	if _, err := tx.Exec(`
		INSERT INTO pending_holds (account_id, amount) VALUES ($1, $2)`, 1, holdAmount); err != nil {
		log.Printf("hold%d: insert =%v", holdAmount, err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("hold %d: Commit =%v", holdAmount, err)
		return
	}

	log.Printf("hold%d: COMMITTED", holdAmount)
}

func main() {

	db, err := sql.Open("postgres", "postgres://localhost/ssl?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	iso := sql.LevelRepeatableRead

	//Concurrency contract:

	chan1 := make(chan struct{}, 1)
	chan2 := make(chan struct{}, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	go worker(&wg, db, ctx, iso, 60, chan1, chan2) //crossed: channel1 is mySignal here...
	go worker(&wg, db, ctx, iso, 60, chan2, chan1) // ...otherSignal there. swap = deadlock

	wg.Wait() // main unlocks when it reads zero on the counter

	var total int
	db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM pending_holds WHERE account_id=1`).Scan(&total)
	fmt.Printf("total holds: %d (balance 100) - invariant %s\n", total,
		map[bool]string{true: "HELD", false: "VIOLATED"}[total <= 100])
}
