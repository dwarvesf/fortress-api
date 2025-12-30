package notion

// PayoutSourceType represents the source type of a payout entry
type PayoutSourceType string

const (
	PayoutSourceTypeContractorPayroll PayoutSourceType = "Contractor Payroll"
	PayoutSourceTypeCommission        PayoutSourceType = "Commission"
	PayoutSourceTypeRefund            PayoutSourceType = "Refund"
	PayoutSourceTypeOther             PayoutSourceType = "Other"
)

// PayoutDirection represents the direction of a payout
type PayoutDirection string

const (
	PayoutDirectionOutgoing PayoutDirection = "Outgoing (you pay)"
	PayoutDirectionIncoming PayoutDirection = "Incoming (you receive)"
)

// PayoutLineItem represents a unified line item for invoice generation
// This aggregates data from ContractorFees, InvoiceSplit, or RefundRequest
type PayoutLineItem struct {
	SourceType  PayoutSourceType
	Direction   PayoutDirection
	Title       string
	Description string
	Hours       float64 // Contractor Payroll only
	Rate        float64 // Contractor Payroll only
	Amount      float64
	AmountUSD   float64
	Currency    string
}

// IsOutgoing returns true if this payout is outgoing (company pays contractor)
func (p *PayoutLineItem) IsOutgoing() bool {
	return p.Direction == PayoutDirectionOutgoing
}

// IsIncoming returns true if this payout is incoming (deduction from contractor)
func (p *PayoutLineItem) IsIncoming() bool {
	return p.Direction == PayoutDirectionIncoming
}

// SignedAmount returns the amount with correct sign based on direction
// Outgoing: positive (company pays)
// Incoming: negative (deduction)
func (p *PayoutLineItem) SignedAmount() float64 {
	if p.Direction == PayoutDirectionIncoming {
		return -p.Amount
	}
	return p.Amount
}

// SignedAmountUSD returns the USD amount with correct sign based on direction
func (p *PayoutLineItem) SignedAmountUSD() float64 {
	if p.Direction == PayoutDirectionIncoming {
		return -p.AmountUSD
	}
	return p.AmountUSD
}
