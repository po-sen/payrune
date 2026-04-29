package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	paymentAddressID, err := parsePositiveInt64Segment(r.PathValue("paymentAddressId"))
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid paymentAddressId")
		return
	}

	response, err := c.getPaymentStatus.Execute(r.Context(), inport.GetPaymentAddressStatusInput{
		Chain:            chain,
		PaymentAddressID: paymentAddressID,
	})
	if err != nil {
		statusCode, message := mapGetPaymentAddressStatusError(err)
		logMappedControllerError(r, statusCode, message, err)
		writeErrorJSON(w, statusCode, message)
		return
	}

	writeJSON(w, http.StatusOK, newPaymentAddressStatusResponse(response))
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

func mapGetPaymentAddressStatusError(err error) (int, string) {
	switch {
	case errors.Is(err, inport.ErrPaymentAddressNotFound):
		return http.StatusNotFound, "payment address is not found"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
