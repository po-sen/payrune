package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
)

type GetPaymentAddressStatusController struct {
	getPaymentStatus inport.GetPaymentAddressStatusUseCase
}

func NewGetPaymentAddressStatusController(
	getPaymentStatus inport.GetPaymentAddressStatusUseCase,
) *GetPaymentAddressStatusController {
	return &GetPaymentAddressStatusController{getPaymentStatus: getPaymentStatus}
}

func (c *GetPaymentAddressStatusController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chain, ok := parseSupportedChainPathValue(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "method not allowed"})
		return
	}

	paymentAddressID, err := parsePositiveInt64Segment(r.PathValue("paymentAddressId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "invalid paymentAddressId"})
		return
	}

	response, err := c.getPaymentStatus.Execute(r.Context(), dto.GetPaymentAddressStatusInput{
		Chain:            chain,
		PaymentAddressID: paymentAddressID,
	})
	if err != nil {
		switch {
		case errors.Is(err, inport.ErrPaymentAddressNotFound):
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func parsePositiveInt64Segment(raw string) (int64, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, errors.New("value must be greater than zero")
	}
	return parsed, nil
}
