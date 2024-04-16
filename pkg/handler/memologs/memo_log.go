package memologs

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/memologs/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

// New returns a handler
func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: controller,
		store:      store,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
	}
}

func (h *handler) Create(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "memologs",
			"method":  "Create",
		},
	)

	body := request.CreateMemoLogsRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		l.Error(err, "[memologs.Create] failed to decode body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	memologs := make([]model.MemoLog, 0)
	for _, b := range body {
		publishedAt, _ := time.Parse(time.RFC3339Nano, b.PublishedAt)
		authors := make([]model.MemoLogAuthor, 0)
		for _, author := range b.Authors {
			discordID, found := whitelistedUsers[author]
			if !found {
				discordMember, err := h.service.Discord.GetMemberByUsername(author)
				if err != nil {
					l.Errorf(err, "[memologs.Create] failed to get discord member", "author", author)
					continue
				}
				discordID = discordMember.User.ID
			}

			github := ""
			employeeID := ""

			if discordID != "" {
				empl, err := h.store.Employee.GetByDiscordID(h.repo.DB(), discordID, true)
				if err != nil {
					l.Errorf(err, "[brainerylogs.Create] failed to get employee with discord id %v", discordID)
				}

				if empl != nil {
					employeeID = empl.ID.String()
					githubSA := model.SocialAccounts(empl.SocialAccounts).GetGithub()
					if githubSA != nil {
						github = githubSA.AccountID
					}
				}
			}

			authors = append(authors, model.MemoLogAuthor{
				DiscordID:  discordID,
				GithubID:   github,
				EmployeeID: employeeID,
			})
		}

		b := model.MemoLog{
			Title:       b.Title,
			URL:         b.URL,
			Authors:     authors,
			Tags:        b.Tags,
			PublishedAt: &publishedAt,
			Description: b.Description,
			Reward:      b.Reward,
		}
		memologs = append(memologs, b)
	}

	logs, err := h.store.MemoLog.Create(h.repo.DB(), memologs)
	if err != nil {
		l.Errorf(err, "[memologs.Create] failed to create new memo logs", "memologs", memologs)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, memologs, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToMemoLog(logs), nil, nil, body, ""))
}

func (h *handler) List(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "memologs",
			"method":  "List",
		},
	)

	memoLogs, err := h.store.MemoLog.List(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get memologs")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToMemoLog(memoLogs), nil, nil, nil, ""))
}

func (h *handler) Sync(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "memologs",
			"method":  "Sync",
		},
	)

	list, err := h.store.MemoLog.List(h.repo.DB())
	if err != nil {
		l.Error(err, "[memologs.Sync] failed to get memo logs")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	mapMemo := make(map[string]bool)
	for _, memo := range list {
		mapMemo[memo.URL] = true
	}

	url := "https://memo.d.foundation/index.xml"
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	d := Rss{}
	err = xml.Unmarshal(data, &d)

	if err != nil {
		log.Fatal(err)
	}

	authorsMap := make(map[string]model.MemoLogAuthor)
	memoLogs := make([]model.MemoLog, 0)
	for _, item := range d.Channel.Item {
		if mapMemo[item.Link] {
			continue
		}

		authorStrList := strings.Split(item.Author, ",")
		memoAuthors := make([]model.MemoLogAuthor, 0)
		for _, author := range authorStrList {
			author = strings.TrimSpace(author)
			if author == "" {
				continue
			}

			discordID, found := whitelistedUsers[author]
			if !found {
				continue
			}

			dt, found := authorsMap[discordID]
			if found {
				memoAuthors = append(memoAuthors, dt)
				continue
			}

			empl, err := h.store.Employee.GetByDiscordID(h.repo.DB(), discordID, true)
			if err != nil {
				l.Error(err, fmt.Sprintf("[memologs.Sync] failed to get employee with discord username %v", discordID))
			}

			github := ""
			discord := discordID
			employeeID := ""
			if empl != nil {
				employeeID = empl.ID.String()
				githubSA := model.SocialAccounts(empl.SocialAccounts).GetGithub()
				if githubSA != nil {
					github = githubSA.AccountID
				}

				if empl.DiscordAccount != nil {
					discord = empl.DiscordAccount.DiscordID
				}
			}

			if github != "" || discord != "" || employeeID != "" {
				memoAuthor := model.MemoLogAuthor{
					GithubID:   github,
					DiscordID:  discord,
					EmployeeID: employeeID,
				}
				authorsMap[discordID] = memoAuthor
				memoAuthors = append(memoAuthors, memoAuthor)
			}
		}

		layout := "Mon, 02 Jan 2006 15:04:05 -0700"
		publishedAt, err := time.Parse(layout, item.PubDate)
		if err != nil {
			l.Error(err, fmt.Sprintf("[memologs.Sync] failed to parse date %v", item.PubDate))
		}

		memoLogs = append(memoLogs, model.MemoLog{
			Title:       item.Title,
			URL:         item.Link,
			Authors:     memoAuthors,
			Description: item.Description,
			PublishedAt: &publishedAt,
			Reward:      decimal.Decimal{},
		})
	}

	if len(memoLogs) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "no new memo logs"))
		return
	}

	results, err := h.store.MemoLog.Create(h.repo.DB(), memoLogs)
	if err != nil {
		l.Error(err, "[memologs.Sync] failed to create memo logs")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToMemoLog(results), nil, nil, nil, "ok"))
}

var whitelistedUsers = map[string]string{
	"thanh":        "790170208228212766",
	"giangthan":    "797051426437595136",
	"nikki":        "796991130184187944",
	"anna":         "575525326181105674",
	"monotykamary": "184354519726030850",
	"vhbien":       "421992793582469130",
	"minhcloud":    "1007496699511570503",
	"mickwan1234":  "383793994271948803",
	"hnh":          "567326528216760320",
	"duy":          "788351441097195520",
	"huytq":        "361172853326086144",
	"han":          "151497832853929986",
	"namtran":      "785756392363524106",
	"innno_":       "753995829559165044",
	"dudaka":       "282500790692741121",
	"tom":          "184354519726030850",
}

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Atom    string   `xml:"atom,attr"`
	Version string   `xml:"version,attr"`
	Script  string   `xml:"script"`
	Channel struct {
		Text  string `xml:",chardata"`
		Title string `xml:"title"`
		Link  struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
		Description    string `xml:"description"`
		Generator      string `xml:"generator"`
		Language       string `xml:"language"`
		ManagingEditor string `xml:"managingEditor"`
		WebMaster      string `xml:"webMaster"`
		Copyright      string `xml:"copyright"`
		LastBuildDate  string `xml:"lastBuildDate"`
		Item           []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			PubDate     string `xml:"pubDate"`
			Author      string `xml:"author"`
			Guid        string `xml:"guid"`
			Description string `xml:"description"`
			Draft       string `xml:"draft"`
		} `xml:"item"`
	} `xml:"channel"`
}
