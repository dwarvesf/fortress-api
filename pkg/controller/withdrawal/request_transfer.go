package withdrawal

import "fmt"

type WithdrawInput struct {
	DiscordID string `json:"discordID"`
	Amount    string `json:"amount"`
}

// RequestTransfer request to transfer token to APP.
func (c *controller) RequestTransfer(in WithdrawInput) (string, error) {
	//l := c.logger.Fields(logger.Fields{
	//	"controller": "withdrawal",
	//	"method":     "Withdraw",
	//})

	profile, err := c.mochiAppClient.GetByDiscordID(in.DiscordID)
	if err != nil {
		return "", err
	}

	fmt.Println(profile)

	return profile.ProfileName, nil
}
