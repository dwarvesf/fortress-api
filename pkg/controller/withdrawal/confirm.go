package withdrawal

type ConfirmTransactionInput struct {
	Timestamp     int64
	TransactionID int64
	Amount        string
	TokenName     string
	ProfileID     string
	RequestCode   string
	Status        string
	Description   string

	BankAccount string
	SwiftCode   string
	Bin         string
}

// ConfirmTransaction means the user has been deposit token to the app.
// The app will transfer the money to then bank account.
func (c *controller) ConfirmTransaction(in ConfirmTransactionInput) error {
	//l := c.logger.Fields(logger.Fields{
	//	"controller": "moneywithdrawal",
	//	"method":     "ConfirmTransaction",
	//})

	return nil
}
