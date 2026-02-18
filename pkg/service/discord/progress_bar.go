package discord

import (
	"github.com/bwmarrin/discordgo"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// Reporter abstracts the mechanism for pushing a Discord embed update.
type Reporter interface {
	Update(embed *discordgo.MessageEmbed) error
}

// ProgressBar wraps a Reporter and handles error logging on failed updates.
type ProgressBar struct {
	reporter Reporter
	logger   logger.Logger
}

// NewProgressBar creates a new ProgressBar backed by the given reporter.
func NewProgressBar(reporter Reporter, l logger.Logger) *ProgressBar {
	return &ProgressBar{
		reporter: reporter,
		logger:   l,
	}
}

// Report sends an embed update (progress or final state) via the underlying reporter.
// Errors are logged but not propagated, so callers can fire-and-forget.
func (pb *ProgressBar) Report(embed *discordgo.MessageEmbed) {
	if err := pb.reporter.Update(embed); err != nil {
		pb.logger.Error(err, "failed to report discord progress update")
	}
}

// channelMessageReporter implements Reporter by editing a channel message.
type channelMessageReporter struct {
	svc       IService
	channelID string
	messageID string
}

// NewChannelMessageReporter creates a Reporter that updates an existing channel message.
// Used for batch invoice progress updates (impl 2).
func NewChannelMessageReporter(svc IService, channelID, messageID string) Reporter {
	return &channelMessageReporter{
		svc:       svc,
		channelID: channelID,
		messageID: messageID,
	}
}

func (r *channelMessageReporter) Update(embed *discordgo.MessageEmbed) error {
	_, err := r.svc.UpdateChannelMessage(r.channelID, r.messageID, "", []*discordgo.MessageEmbed{embed}, nil)
	return err
}

// interactionReporter implements Reporter by editing the original interaction response.
// Used for extra payment notification progress updates (impl 3).
type interactionReporter struct {
	svc   IService
	appID string
	token string
}

// NewInteractionReporter creates a Reporter that edits an ephemeral interaction response.
func NewInteractionReporter(svc IService, appID, token string) Reporter {
	return &interactionReporter{
		svc:   svc,
		appID: appID,
		token: token,
	}
}

func (r *interactionReporter) Update(embed *discordgo.MessageEmbed) error {
	return r.svc.EditInteractionResponse(r.appID, r.token, []*discordgo.MessageEmbed{embed})
}
