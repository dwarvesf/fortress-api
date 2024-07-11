package earn

import (
	"context"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

const memoURL = "https://memo.d.foundation"

func (c *controller) ListEarn(ctx context.Context) ([]model.Earn, error) {
	htmlContent, err := fetchHTML(memoURL + "/earn")
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	earns := parseEarnTable(doc)

	picIDs := make([]string, 0)
	for _, earn := range earns {
		picIDs = append(picIDs, earn.PICs...)
	}

	pics, err := c.store.DiscordAccount.ListByMemoUsername(c.repo.DB(), picIDs)
	if err != nil {
		return nil, err
	}

	picDiscordIDMap := make(map[string]string)
	for _, pic := range pics {
		key := pic.MemoUsername
		if key == "" {
			key = pic.DiscordUsername
		}

		picDiscordIDMap[key] = pic.DiscordID
	}

	for i, earn := range earns {
		for j, picID := range earn.PICs {
			if val, ok := picDiscordIDMap[picID]; ok {
				earn.PICs[j] = val
			}
		}
		earns[i] = earn
	}

	return earns, nil
}

func fetchHTML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func parseEarnTable(node *html.Node) []model.Earn {
	var earns []model.Earn
	var currentEarn model.Earn
	var isHeader bool
	var fieldIndex int

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			isHeader = false
			fieldIndex = 0
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "th" {
					isHeader = true
					break
				}
			}
			if !isHeader {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "td" {
						text := getText(c)
						switch fieldIndex {
						case 0:
							if linkNode := findLinkNode(c); linkNode != nil {
								currentEarn.URL = memoURL + getAttr(linkNode, "href")
								currentEarn.Title = getText(linkNode)
							} else {
								currentEarn.Title = text
							}
						case 1:
							currentEarn.Bounty = text
						case 2:
							currentEarn.Status = text
						case 3:
							pics := strings.Split(text, ",")
							for _, pic := range pics {
								if pic != "" {
									currentEarn.PICs = append(currentEarn.PICs, pic)
								}
							}
						case 4:
							currentEarn.Function = text
						}
						fieldIndex++
					}
				}
				earns = append(earns, currentEarn)
				currentEarn = model.Earn{}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(node)

	return earns
}

func findLinkNode(n *html.Node) *html.Node {
	var linkNode *html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			linkNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return linkNode
}

func getAttr(n *html.Node, attrName string) string {
	for _, attr := range n.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

func getText(n *html.Node) string {
	var buf strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return buf.String()
}
