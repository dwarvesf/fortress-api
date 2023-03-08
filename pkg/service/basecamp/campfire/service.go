package campfire

type Service interface {
	CreateLine(projectID int, campfireID int, line string) (err error)
	BotCreateLine(projectID int, campfireID int, line string) (err error)
	BotReply(callbackURL string, message string) (err error)
}
