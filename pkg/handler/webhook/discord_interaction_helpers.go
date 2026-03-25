package webhook

import (
	"crypto/ed25519"
	"encoding/hex"
	"net/http"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/leekchan/accounting"
)

// verifyDiscordSignature verifies the Discord interaction signature
func verifyDiscordSignature(publicKey, signature, timestamp string, body []byte) bool {
	// Decode public key
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false
	}

	// Decode signature
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	// Create message to verify
	message := append([]byte(timestamp), body...)

	// Verify signature
	return ed25519.Verify(pubKeyBytes, message, sigBytes)
}

// respondToInteraction sends a simple response to an interaction
func (h *handler) respondToInteraction(c *gin.Context, message string) {
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral, // Only visible to the user
		},
	}
	c.JSON(http.StatusOK, response)
}

// formatCurrency formats an amount as USD currency with comma separators
// e.g., 2290.26 -> "$2,290.26", 1000 -> "$1,000"
func formatCurrency(amount float64) string {
	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	// If whole number, don't show decimals
	if amount == float64(int(amount)) {
		ac.Precision = 0
	}
	return ac.FormatMoney(amount)
}

// stripInvoiceIDFromReason removes invoice ID from reason format
// e.g., "[RENAISS :: INV-DO5S8] ..." → "[RENAISS] ..."
func stripInvoiceIDFromReason(reason string) string {
	re := regexp.MustCompile(`\[([^\]]+?)\s*::\s*INV-[A-Z0-9]+\]`)
	return re.ReplaceAllString(reason, "[$1]")
}
