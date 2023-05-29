package googleadmin

// IService interface contain related google calendar method
type IService interface {
	GetGroupMemberEmails(groupEmail string) ([]string, error)
	DeleteAccount(mail string) error
}
