package brainerylogs

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/brainerylogs/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/brainerylogs/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/view"
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

const (
	braineryLogsChannelID           = "955015316293972048"
	braineryLogsPlaygroundChannelID = "1119171172198797393"
)

// Create godoc
// @Summary Create brainery logs
// @Description Create brainery logs
// @id createBraineryLog
// @Tags Project
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param body body CreateBraineryLogRequest true "Body"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /brainery-logs [post]
func (h *handler) Create(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "brainerylogs",
			"method":  "Create",
		},
	)

	body := request.CreateBraineryLogRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		l.Error(err, "failed to decode body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}
	if err := body.Validate(); err != nil {
		l.Errorf(err, "failed to validate data", "body", body)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	publishedAt, _ := time.Parse(time.RFC3339Nano, body.PublishedAt)

	b := model.BraineryLog{
		Title:       body.Title,
		URL:         body.URL,
		GithubID:    body.GithubID,
		DiscordID:   body.DiscordID,
		Tags:        body.Tags,
		PublishedAt: &publishedAt,
		Reward:      body.Reward,
	}

	log, err := h.controller.BraineryLog.Create(b)
	if err != nil {
		l.Errorf(err, "failed to create brainery logs", "braineryLog", b)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, log, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any]("success", nil, nil, body, ""))
}

// GetMetrics godoc
// @Summary Get brainery metric
// @Description Get brainery metric
// @id getBraineryMetric
// @Tags Project
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param view query string false "Time view"
// @Param date query string false "Date" Format(date)
// @Success 200 {object} BraineryMetricResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /brainery-logs/metrics [get]
func (h *handler) GetMetrics(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "brainerylogs",
			"method":  "GetMetrics",
		},
	)

	queryView := c.DefaultQuery("view", "weekly")
	date := c.Query("date")
	selectedDate := time.Now()

	if date != "" {
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			l.Error(err, "failed to parse date")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidDateFormat, nil, ""))
			return
		}

		selectedDate = t
	}

	latestPosts, logs, ncids, err := h.controller.BraineryLog.GetMetrics(selectedDate, queryView)
	if err != nil {
		l.Error(err, "failed to get brainery metrics")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToBraineryMetric(latestPosts, logs, ncids, queryView), nil, nil, nil, ""))
}

func (h *handler) Sync(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "brainerylogs",
			"method":  "Sync",
		},
	)

	githubPattern := `github\.com/([\w-]+)`
	githubRe := regexp.MustCompile(githubPattern)

	displayNameIDMap := map[string]string{
		"An Tran":           "656dd867-39c1-4ed8-b72e-63d0e89c3679",
		"Bach Phuong":       "1782aaea-733c-4ced-8d30-a921abede14e",
		"Bien Vo":           "5a8f3d89-04a0-4a0e-bdc8-73045ced6a08",
		"Binh Le":           "06126b41-b20f-4ca2-933a-06fdb01fe362",
		"Dung Ho":           "bf1d3fcc-6a59-413e-a692-113989849556",
		"Hieu Phan":         "e0d8c920-f4f1-4309-aabf-1d83f792f45b",
		"Huy Tieu":          "3c420751-bb9a-4878-896e-2f10f3a633d6",
		"Khanh Truong":      "63d163a7-e9f5-4210-a685-151061fe9c29",
		"Khoi Nguyen":       "224b9b85-7206-46bd-803a-5308fd4e81e1",
		"Lap Nguyen":        "e84c1860-6991-44a1-b687-fec078b842eb",
		"Le Duc Chinh":      "40f33d74-91ae-4eec-ba35-8d330376d6e1",
		"M.Vu Cuong(Jim)":   "07a5a7b2-6e5f-4165-bbdf-40d06f3e7837",
		"Ngo Lap Nguyen":    "e84c1860-6991-44a1-b687-fec078b842eb",
		"Ngo Trong Khoi":    "8281ade3-214a-4348-a171-42b0fe304032",
		"Nguyen Dinh Nam":   "a5ddbb54-3faa-4f92-964a-3754928d3f21",
		"Nguyen Huu Nguyen": "9a91ff6b-7367-403f-b23d-4c16dabd6857",
		"Nguyen Tran Khanh": "14bdccb9-3460-40f6-ba87-6ad1a7884670",
		"Nguyen Xuan Anh":   "69bf5adf-7ba2-4abc-b87e-9a68668a267e",
		"Nhut Huynh":        "a8f34385-61f8-46f0-9403-4b05e37cd8e3",
		"Pham Duc Thanh":    "3e858c81-d661-4d4f-b913-e02dd6f4007e",
		"Pham Ngoc Thanh":   "ec086e4a-2167-4924-adfd-84be02edebe9",
		"Pham The Hung":     "133483cd-7a76-4a6d-9c63-964900284a44",
		"Phan Viet Trung":   "ee219b4e-e4dc-4782-b590-03f799cd41ab",
		"Phat Ha":           "1b63c305-b04b-47ab-ab1b-e6810d8a17cc",
		"Phuc Le":           "d8a6af04-9e0a-4724-97c4-a78ecd5e9bc4",
		"Thanh Pham":        "3e858c81-d661-4d4f-b913-e02dd6f4007e",
		"Tom Nguyen":        "69bf5adf-7ba2-4abc-b87e-9a68668a267e",
		"Tran Hoang Nam":    "d8b2785c-31f1-451b-8bb5-385cbc873402",
		"Tran Khac Vy":      "2641297b-c632-4c22-9a42-61291d621552",
		"Trung Phan":        "ee219b4e-e4dc-4782-b590-03f799cd41ab",
		"Truong Hung Khanh": "63d163a7-e9f5-4210-a685-151061fe9c29",
		"Vy Tran":           "2641297b-c632-4c22-9a42-61291d621552",
	}

	body := request.SyncBraineryLogs{}
	if err := c.ShouldBindJSON(&body); err != nil {
		l.Errorf(err, "failed to parse request")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if body.StartMessageID == "" {
		body.StartMessageID = "960194750806364171"
	}

	if body.EndMessageID == "" {
		body.EndMessageID = "1120663834261205052"
	}

	channelID := braineryLogsChannelID
	if h.config.Env != "prod" {
		channelID = braineryLogsChannelID
	}

	messages, err := h.service.Discord.GetMessagesAfterCursor(channelID, body.StartMessageID, body.EndMessageID)
	if err != nil {
		l.Errorf(err, "failed to get messages from discord", "startMessageID", body.StartMessageID, "endMessageID", body.EndMessageID)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	socialAccounts, err := h.store.SocialAccount.GetByType(h.repo.DB(), model.SocialAccountTypeGitHub.String())
	if err != nil {
		l.Errorf(err, "failed to get social accounts")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	githubEmployeeIDMap, empIDGithubMap := model.SocialAccounts(socialAccounts).ToMap()

	employees, _, err := h.store.Employee.All(h.repo.DB(), employee.EmployeeFilter{}, model.Pagination{
		Page: 0,
		Size: 1000,
	})
	if err != nil {
		l.Errorf(err, "failed to get employees")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	employeeDiscordMap := make(map[string]string)
	for _, e := range employees {
		if e.DiscordAccount != nil && e.DiscordAccount.DiscordID != "" {
			employeeDiscordMap[e.ID.String()] = e.DiscordAccount.DiscordID
		}
	}

	var braineryLogs []model.BraineryLog
	for _, m := range messages {
		if !m.Author.Bot {
			continue
		}

		if len(m.Embeds) == 0 {
			continue
		}

		githubID := ""
		discordID := ""
		employeeID := ""
		var tags []string
		for _, f := range m.Embeds[0].Fields {
			if f.Name == "Topic" {
				if f.Value != "" {
					tmp := strings.ReplaceAll(f.Value, "#", "")
					tmp = strings.ReplaceAll(tmp, "\n", "/")
					tmp = strings.ReplaceAll(tmp, " ", "")
					tmp = strings.ReplaceAll(tmp, "...", "")
					tmp = strings.ToLower(tmp)
					tmpTags := strings.Split(tmp, "/")
					tags = append(tags, tmpTags...)
				}
			}

			if f.Name == "Tags" {
				if f.Value != "" {
					tmp := strings.ReplaceAll(f.Value, "#", "")
					tmp = strings.ReplaceAll(tmp, "\n", "/")
					tmp = strings.ReplaceAll(tmp, " ", "")
					tmp = strings.ReplaceAll(tmp, "...", "")
					tmp = strings.ToLower(tmp)
					tmpTags := strings.Split(tmp, "/")
					tags = append(tags, tmpTags...)
				}
			}

			if f.Name == "Author" && strings.Contains(m.Embeds[0].URL, "https://github.com/dwarvesf") {
				githubID = f.Value
				v, ok := githubEmployeeIDMap[f.Value]
				if ok {
					employeeID = v
					d, ok := employeeDiscordMap[v]
					if ok {
						discordID = d
					}
				}
			}

			if f.Name == "Author" && strings.Contains(m.Embeds[0].URL, "https://brain.d.foundation") {
				github := githubRe.FindStringSubmatch(f.Value)
				if len(github) > 1 {
					githubID = github[1]
					v, ok := githubEmployeeIDMap[github[1]]
					if ok {
						employeeID = v
						d, ok := employeeDiscordMap[v]
						if ok {
							discordID = d
						}
					}
				} else {
					v, ok := displayNameIDMap[f.Value]
					if ok {
						employeeID = v
						d, ok := employeeDiscordMap[v]
						if ok {
							discordID = d
						}

						gh, ok := empIDGithubMap[v]
						if ok {
							githubID = gh
						}
					} else {
						githubID = f.Value
					}
				}
			}
		}

		eid := model.UUID{}
		euuid, err := model.UUIDFromString(employeeID)
		if err == nil {
			eid = euuid
		}

		loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
		t := m.Timestamp.In(loc)
		l := model.BraineryLog{
			Title:       m.Embeds[0].Title,
			URL:         m.Embeds[0].URL,
			GithubID:    githubID,
			DiscordID:   discordID,
			EmployeeID:  eid,
			Tags:        tags,
			PublishedAt: &t,
			Reward:      decimal.NewFromInt(10),
		}

		braineryLogs = append(braineryLogs, l)
	}

	if len(braineryLogs) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
		return
	}

	_, err = h.store.BraineryLog.Create(h.repo.DB(), braineryLogs)
	if err != nil {
		l.Errorf(err, "failed to create brainery logs")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
