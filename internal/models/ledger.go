package models

impot "time"

type TransactionStatus string

const (
	StatusPending TransactionStatus = "PENDING"
	StatusSettled TransactionStatus = "SETTLED"
	StatusFailed TransactionStatus = "FAILED"
)


type Transaction struct {

	ID string `json:"id"`
	DebitAccount string `json:"debit_account_id"`
	CreditAccount string `json:credit_account_id"`
	Amount int64 `json:"amount"`
	Currency string `json:"currency"`
	Metadata map[string]string `json:"metadata"`
	Timestamp time.Time `json:"timestamp"`

}

type SettlementResult struct {
	
	TransactionID string
	Status TransactionStatus
	Message string
	ProcessedBy int
}