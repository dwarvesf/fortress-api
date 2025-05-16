package memologs

// TODO: deprecated, now memo have no author - 16 May 2025
func (c *controller) Sweep() error {
	c.logger.Warn("deprecated, now memo have no author - 16 May 2025")
	return nil

	// List memo logs without authors
	// memoLogs, err := c.store.MemoLog.ListNonAuthor(c.repo.DB())
	// if err != nil {
	// 	return err
	// }

	// // Fetch memo data from the RSS feed
	// resp, err := http.Get(dfMemoRssURL)
	// if err != nil {
	// 	return err
	// }
	// defer resp.Body.Close()

	// var feed struct {
	// 	Items []struct {
	// 		Link   string `xml:"link"`
	// 		Author string `xml:"author"`
	// 	} `xml:"channel>item"`
	// }

	// if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
	// 	return err
	// }

	// memoData := make(map[string][]string)
	// for _, item := range feed.Items {
	// 	authors := strings.Split(item.Author, ",")
	// 	for i := 0; i < len(authors); i++ {
	// 		authors[i] = strings.TrimSpace(authors[i])
	// 		if authors[i] == "" {
	// 			authors = append(authors[:i], authors[i+1:]...)
	// 			i--
	// 		}
	// 	}

	// 	if len(authors) == 0 {
	// 		continue
	// 	}
	// 	memoData[item.Link] = authors
	// }

	// // Process each memo log
	// for _, memoLog := range memoLogs {
	// 	usernames := memoData[memoLog.URL]
	// 	if len(usernames) == 0 {
	// 		continue
	// 	}

	// 	// Fetch Discord accounts for the usernames
	// 	discordAccounts, err := c.store.DiscordAccount.ListByMemoUsername(c.repo.DB(), usernames)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// Create memo authors
	// 	for _, discordAccount := range discordAccounts {
	// 		memoAuthor := &model.MemoAuthor{
	// 			MemoLogID:        memoLog.ID,
	// 			DiscordAccountID: discordAccount.ID,
	// 		}
	// 		if err := c.store.MemoLog.CreateMemoAuthor(c.repo.DB(), memoAuthor); err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	// return nil
}
