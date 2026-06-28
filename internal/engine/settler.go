package engine

import (
	"context"
	"log/slog"

	"github.com/Soufiane1412/sovereign-ledger-go/internal/models"
)

// StartSettler acts as a sovereign node processing trades
func StartSettler(ctx context.Context, nodeID int, jobs <-chan models.Transaction, results chan<- models.SettlementResult, logger *slog.Logger) {

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		select {
		case <-ctx.Done():
			return
		case tx, ok := <-jobs:
			if !ok {
				return
			}
			result := models.SettlementResult{
				TransactionID: tx.ID,
				Status:        models.StatusFailed,
				Message:       "SAMA Audit required: High value Cross-Border Trade",
				ProcessedBy:   nodeID,
			}

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}

	}
}
