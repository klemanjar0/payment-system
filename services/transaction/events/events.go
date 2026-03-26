package events

// Events published by the transaction service (consumed by account service).
const (
	HoldRequested        = "transaction.hold_requested"
	ExecuteHoldRequested = "transaction.execute_hold_requested"
	ReleaseHoldRequested = "transaction.release_hold_requested"
	CreditRequested      = "transaction.credit_requested"
)

// Events published by the transaction service (internal lifecycle).
const (
	TransactionInitiated = "transaction.initiated"
	TransactionCompleted = "transaction.completed"
	TransactionFailed    = "transaction.failed"
)

type HoldRequestedPayload struct {
	AccountID     string `json:"account_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
}

type ExecuteHoldRequestedPayload struct {
	AccountID     string `json:"account_id"`
	TransactionID string `json:"transaction_id"`
}

type ReleaseHoldRequestedPayload struct {
	AccountID     string `json:"account_id"`
	TransactionID string `json:"transaction_id"`
}

type CreditRequestedPayload struct {
	AccountID     string `json:"account_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
}
