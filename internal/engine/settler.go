package engine

import (
	"fmt"
	"time"

	"github.com/Soufiane1412/sovereign-ledger-go/internal/models"
)

// StartSettler acts as a sovereign node processing trades
func StartSettler(nodeID int, jobs <-chan models.Transaction, results chan<- models.SettlementResult) {

	// This loop blocks until a job is pushed into the 'jobs' channel
	for tx := range jobs {

		// 1. SAMA Compliance Filter (Business Logic)
		// Halalas: 100,000,000 = 1,000,000,000 SAR, High-value trades require manual
		if tx.Amount > 100000000 {
			results <- models.SettlementResult{

				TransactionID: tx.ID,
				Status:        models.StatusFailed,
				Message:       "SAMA Audit required: High value Cross-Border Trade",
				ProcessedBy:   nodeID,
			}
			continue
		}

		// 2. The Atomic Operation Simulation
		// In a real system this where we lock the DB rows for Debit and Credit
		time.Sleep(10 * time.Millisecond)

		// Log the successful settlement for the audit trail
		fmt.Printf("[Sovereign Node %d] Executed: %.2f %s | %s -> %s\n", nodeID, float64(tx.Amount)/100, tx.Currency, tx.DebitAccount, tx.CreditAccount)

		// 3. Output the result to the synchronisation channel
		results <- models.SettlementResult{
			TransactionID: tx.ID,
			Status:        models.StatusSettled,
			Message:       "Trade Settled Locally",
			ProcessedBy:   nodeID,
		}
	}
}
