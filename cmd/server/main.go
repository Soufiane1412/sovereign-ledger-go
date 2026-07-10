package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Soufiane1412/sovereign-ledger-go/internal/engine"
	"github.com/Soufiane1412/sovereign-ledger-go/internal/models"
)

func main() {

	// Constants: System-level tuning

	const (
		numNodes      = 100              // Total concurrent workers
		tradeLoad     = 10000            // Total transactions to simulate
		jobBuffer     = 256              // See Ba kpressure note below
		resultBuffer  = 256              // Independant downstream backpressure
		shutdownGrace = 20 * time.Second // Max wait for graceful exit
	)

	//

	// ──────────────────────────────────────────────────────────────────
	// OBSERVABILITY — Structured logging
	// ──────────────────────────────────────────────────────────────────
	// 🧠 fmt.Println is for tutorials. slog (Go 1.21+) produces JSON logs
	// that Datadog/Loki/CloudWatch can parse. SAMA audit trails REQUIRE this.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("system_boot", "nodes", numNodes, "load", tradeLoad)

	// ──────────────────────────────────────────────────────────────────
	// 3. CONTEXT — Universal cancellation token for the whole process tree
	// ──────────────────────────────────────────────────────────────────
	// 🧠 Every goroutine in this system listens to ctx.Done(). Cancel ctx →
	// every worker exits cleanly. This is the SINGLE most important pattern
	// senior Go interviewers test. Memorize the shape.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Belt-and-braces: if main returns early, cancel propagates#

	// ──────────────────────────────────────────────────────────────────
	// 4. SIGNAL HANDLING — Graceful shutdown on SIGTERM/SIGINT
	// ──────────────────────────────────────────────────────────────────
	// 🧠 In Kubernetes, the orchestrator sends SIGTERM before killing the pod.
	// Ignore it → you crash mid-settlement → SAMA auditor screams. Catching
	// SIGTERM and cancelling ctx lets workers drain cleanly before exit.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Warn("shutdown_signal", "signal", sig.String())
		cancel()
	}()

	// 5. Initialise Channels (The 'Pipes')

	jobs := make(chan models.Transaction, jobBuffer)
	results := make(chan models.SettlementResult, resultBuffer)
	// 🧠 OWNERSHIP RULE (memorize for interviews):
	//   • Producer (step 8) owns the SEND-side of jobs → producer closes jobs
	//   • Workers own the SEND-side of results → closer goroutine (step 7) closes
	//   • Nobody else closes either. Violating this = "panic: send on closed channel"

	// ──────────────────────────────────────────────────────────────────
	// 6. FAN-OUT — Spawn worker pool with WaitGroup tracking
	// ──────────────────────────────────────────────────────────────────

	var wg sync.WaitGroup
	for w := 1; w <= numNodes; w++ {

		wg.Add(1) // 🧠 ALWAYS Add BEFORE the `go` statement, never inside.
		// Race condition if inside: main could reach wg.Wait() before any
		// Add ran → Wait returns immediately → buggy "instant completion".
		go func(nodeID int) {
			defer wg.Done()
			engine.StartSettler(ctx, nodeID, jobs, results, logger)

		}(w)
	}

	logger.Info("Workers_started", "count", numNodes)

	// ──────────────────────────────────────────────────────────────────
	// 7. CLOSER GOROUTINE — The graceful-shutdown idiom
	// ──────────────────────────────────────────────────────────────────
	// 🧠 THIS is the load-bearing pattern. Without it, you have two bad options:
	//   (a) never close results → consumer hangs forever
	//   (b) close results too early → worker panics sending to closed channel
	// The closer goroutine bridges main's knowledge ("WaitGroup done?") with
	// the consumer's need ("is the channel closed?"). Memorize this shape.
	go func() {
		wg.Wait()
		close(results)
		logger.Info("results_cvhannel_closed")
	}()

	// ──────────────────────────────────────────────────────────────────
	// 8. PRODUCER — Inject the transaction load
	// ──────────────────────────────────────────────────────────────────
	go func() {
		defer close(jobs)
		for i := 1; i <= tradeLoad; i++ {
			tx := models.Transaction{
				ID:             fmt.Sprintf("TXN-SA-%d", i),
				IdempotencyKey: fmt.Sprintf("idem-%d", i), // 🧠 SAMA: retry-safety key
				DebitAccount:   "FR-MERC-88",
				CreditAccount:  "SA-SUPP-01",
				Amount:         int64(i * 100000), // 🧠 100k→1B halalas; SAMA filter will fire
				Currency:       "SAR",
				Timestamp:      time.Now(),
			}
			// 🧠 SELECT-FOR-CANCELLATION: don't blindly send. If ctx is cancelled
			// while workers are slow and the channel is full, blind send would
			// deadlock forever. select lets you abort cleanly.
			select {
			case jobs <- tx:

			case <-ctx.Done():
				logger.Warn("producer_cancelled", "sent", i, "of", tradeLoad)
				return

			}
		}
		logger.Info("producer_done", "total", tradeLoad)
	}()

	// ──────────────────────────────────────────────────────────────────
	// 9. CONSUMER (FAN-IN) — Aggregate results
	// ──────────────────────────────────────────────────────────────────
	// 🧠 `for range` over results, NOT `for i := 0; i < tradeLoad; i++`.
	// Range exits naturally when the channel closes. Count-based loops are
	// fragile: if one worker dies silently, the count is wrong → main hangs.
	// Range + closer-goroutine = robust under partial failure.

	type tally struct{ settled, failed int }
	done := make(chan tally, 1)
	go func() {
		var t tally
		for res := range results {
			switch res.Status {
			case models.StatusSettled:
				t.settled++
			case models.StatusFailed:
				t.failed++
				logger.Warn("audit_alert",
					"tx_id", res.TransactionID,
					"reason", res.Message,
					"node", res.ProcessedBy)
			}
		}
		done <- t
	}()

	select {
	case t := <-done:
		logger.Info("settlement_complete", "settled", t.settled, "failed", t.failed, "total", t.settled+t.failed)
	case <-time.After(shutdownGrace):
		logger.Error("shutdown_grace_exceeded", "grace", shutdownGrace)

	}

}
