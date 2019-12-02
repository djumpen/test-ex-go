package api

import (
	"context"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/djumpen/test-ex-go/apperrors"
	"github.com/djumpen/test-ex-go/models"
	"github.com/gin-gonic/gin"
)

type StateResultEvent struct {
	State         string `json:"state" binding:"required,oneof=win loss"`
	Amount        string `json:"amount" binding:"required"` // TODO: create numstring validator
	TransactionID string `json:"transactionId" binding:"required"`
}

// ----------------------------------

type eventsService interface {
	Create(context.Context, models.Event) error
}

type eventsResource struct {
	svc  eventsService
	resp SimpleResponder
}

// NewEventsResource returns Events API resource
func NewEventsResource(svc eventsService, resp SimpleResponder) *eventsResource {
	return &eventsResource{
		svc:  svc,
		resp: resp,
	}
}

func (r StateResultEvent) validateToModel() (models.Event, error) {
	amount, err := strconv.ParseFloat(r.Amount, 64)
	if err != nil {
		return models.Event{}, apperrors.NewValidation("request", errors.New("Amount is not valid"))
	}

	state := models.EventState(strings.ToUpper(r.State))

	if state == models.StateLoss && amount > 0 {
		return models.Event{}, apperrors.NewValidation("request", errors.New("Amount is not valid"))
	}
	if state == models.StateWin && amount < 0 {
		return models.Event{}, apperrors.NewValidation("request", errors.New("Amount is not valid"))
	}

	return models.Event{
		State:         state,
		Amount:        amount,
		TransactionID: r.TransactionID,
	}, nil
}

// ProcessNewEvent processes incomig event
func (r *eventsResource) ProcessNewEvent(c *gin.Context) {
	var req StateResultEvent
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.WithStack(err))
		return
	}
	event, err := req.validateToModel()
	if err != nil {
		c.Error(errors.WithStack(err))
		return
	}
	err = r.svc.Create(c, event)
	if err != nil {
		c.Error(errors.WithStack(err))
		return
	}
	r.resp.Created(c, gin.H{
		"transactionId": event.TransactionID,
	})
}
