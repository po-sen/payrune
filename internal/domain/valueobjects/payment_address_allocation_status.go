package valueobjects

type PaymentAddressAllocationStatus string

const (
	PaymentAddressAllocationStatusReserved         PaymentAddressAllocationStatus = "reserved"
	PaymentAddressAllocationStatusIssued           PaymentAddressAllocationStatus = "issued"
	PaymentAddressAllocationStatusDerivationFailed PaymentAddressAllocationStatus = "derivation_failed"
)
