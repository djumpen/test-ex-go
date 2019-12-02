package models

type EventStatus string
type EventState string

const (
	StatusProcessed EventStatus = "PROCESSED"
	StatusCanceled  EventStatus = "CANCELED"

	StateWin  EventState = "WIN"
	StateLoss EventState = "LOSS"
)

type Event struct {
	ID            int
	State         EventState
	Amount        float64
	TransactionID string
	Status        EventStatus
}
