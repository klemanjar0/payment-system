package events

const (
	HoldPlaced   = "account.hold.placed"
	HoldExecuted = "account.hold.executed"
	HoldReleased = "account.hold.released"
	HoldFailed   = "account.hold.failed"
	Credited     = "account.credited"
)

type HoldPlacedPayload struct {
	AccountID     string `json:"account_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
}

type HoldExecutedPayload struct {
	AccountID     string `json:"account_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
}

type HoldReleasedPayload struct {
	AccountID     string `json:"account_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
}

type CreditedPayload struct {
	AccountID     string `json:"account_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
}

type HoldFailedPayload struct {
	AccountID     string `json:"account_id"`
	TransactionID string `json:"transaction_id"`
	Reason        string `json:"reason"`
}
