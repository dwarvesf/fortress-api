package improvmx

type IService interface {
	CreateAccount(email, fwdEmail string) error
	DeleteAccount(email string) error
}
