package discord

var mapEmoji = map[string]string{
	"ARROW_DOWN_ANIMATED": "<a:arrow_down_animated:1131789144759214171>",
	"ARROW_UP_ANIMATED":   "<a:arrow_up_animated:1131789319644921936>",
	"BADGE1":              "<a:badge1:1131850989062852638>",
	"BADGE2":              "<a:badge2:1131850991663337554>",
	"BADGE3":              "<a:badge3:1131850996159610930>",
	"BADGE5":              "<a:badge5:1131851001117294672>",
	"LOG_CHANNEL":         "<:log_channel:1131863319377100841>",
	"STAR_ANIMATED":       "<a:star_animated:1131862886592024586>",
	"INCREASING_ANIMATED": "<a:increasing_animated:1131862879319097394>",
	"CLOCK_NEW":           "<:clock_new:1131863089185292428>",
}

func getEmoji(emoji string) string {
	return mapEmoji[emoji]
}
