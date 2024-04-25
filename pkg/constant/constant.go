package constant

const (
	RegexPatternDiscordChannelID = `<#(\d+)>`
	RegexPatternDiscordID        = `<@(\d+)>`
	RegexPatternEmail            = `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]+\b`
	RegexPatternIcyReward        = ` (\d+)`
	RegexPatternNumber           = `\d{18,}`
	RegexPatternUrl              = `((?:https?://)[^\s]+)`
	RegexPatternGithub           = `gh:(\w+)`
	RegexPatternDescription      = `d:"(.*?)"`
	RegexPatternTime             = `t:(\w+)`
)
