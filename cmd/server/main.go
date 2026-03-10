package main

import (
	"fmt"
	"time"

	"github.com/Soufiane1412/sovereign-ledger-go/internal/engine"
	"github.com/Soufiane1412/sovereign-ledger-go/internal/models"
)

func main() {

	// Constants: System-level tuning

	const (
		numNodes  = 10   // Total concurrent workers
		tradeLoad = 1000 // Total transactions to simulate
	)

	// 1. Initialise Channels (The 'Pipes')

	jobs := make(chan models.Transaction, tradeLoad)
	results := make(chan models.SettlementResult, tradeLoad)

	fmt.Println("System Boot: Initialising Sovereign Settlement Nodes...")

	// 2. the Fan-Out: Booting the workers
	for w := 1; w <= numNodes; w++ {

		go engine.StartSettler(w, jobs, results)
	}

	// 3. Data Injection: Simulating a burst of Saudi-EU trade
	go func() {
		for i := 1; i <= tradeLoad; i++ {
			jobs <- models.Transaction{
				ID:            fmt.Sprintf("TXN-SA-%d", i),
				DebitAccount:  "FR-MERC-88",
				CreditAccount: "SA-SUPP-01",
				Amount:        int64(i * 1000), // In Halalas
				Currency:      "SAR",
				Timestamp:     time.Now(),
			}
		}
		// Critical: CLose the pipe so workers know no more work is coming
		close(jobs)

	}()

	// 4. Synchronisation Point: Gathering the audit results
	for a := 1; a <= tradeLoad; a++ {
		res := <-results
		if res.Status == models.StatusFailed {
			fmt.Printf("[AUDIT ALERT] transaction %s Failed: %s\n", res.TransactionID, res.Message)
		}
	}
	fmt.Printf("\n[FINISH] Sovereign Ledger succesfully settled %d transactions. \n", tradeLoad)
}
