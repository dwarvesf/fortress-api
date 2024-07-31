package youtube

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/git"
	"github.com/dwarvesf/fortress-api/pkg/store"
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

func (h *handler) TranscribeBroadcast(c *gin.Context) {
	broadcast, err := h.service.Youtube.GetLatestBroadcast()
	if err != nil {
		h.logger.Error(err, "failed to get latest broadcast")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	link := "https://www.youtube.com/live/" + broadcast.Id

	broadCastTime := strings.Split(broadcast.Snippet.ActualStartTime, "T")
	broadCastDate := time.Now().Format("2006-01-02")
	if len(broadCastTime) == 2 {
		broadCastDate = broadCastTime[0]
		broadCastDate = strings.ReplaceAll(broadCastDate, "-", "")
	}

	content, err := h.service.Dify.SummarizeOGIFMemo(link)
	if err != nil {
		h.logger.Error(err, "failed to summarize OGIF memo")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	gitSvc := git.New("https://github.com/dwarvesf/brainery", "lmquang", h.config.Github.Token)
	branch := fmt.Sprintf("docs/ogif-memo-summary-%v", time.Now().Format("20060102"))
	if err := gitSvc.CreateBranch(branch); err != nil {
		h.logger.Error(err, "failed to create branch")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	fileNum := 1
	// check the latest file
	path := gitSvc.Dest() + "/updates/ogif"
	// walk through the directory to get the latest file
	// if the file is not exist, create the first file
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	} else {
		files, err := os.ReadDir(path)
		if err != nil {
			h.logger.Error(err, "failed to read directory")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
		// get the latest file number follow this format: 1-ogif-office-hours-20210101.md
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			name := file.Name()
			if strings.Contains(name, "ogif-office-hours") {
				num := strings.Split(name, "-")[0]
				if n, err := strconv.Atoi(num); err == nil {
					if n > fileNum {
						fileNum = n
					}
				}
			}
		}
		fileNum++
	}

	if err := h.createFile(gitSvc.Dest()+"/updates/ogif", fmt.Sprintf("%v-ogif-office-hours-%v.md", fileNum, broadCastDate), content); err != nil {
		h.logger.Error(err, "failed to create file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if err := gitSvc.Commit(fmt.Sprintf("docs: ogif memo summary %v", broadCastDate)); err != nil {
		h.logger.Error(err, "failed to commit")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	if err := gitSvc.Push(); err != nil {
		h.logger.Error(err, "failed to push")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	prNumber, err := gitSvc.CreatePullRequest("dwarvesf", "brainery", branch, "main", fmt.Sprintf("docs: ogif memo summary %v", broadCastDate), "")
	if err != nil {
		h.logger.Error(err, "failed to create PR")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	if err := gitSvc.RequestReview(
		"dwarvesf",
		"brainery",
		*prNumber,
		h.config.Github.BraineryReviewers,
	); err != nil {
		h.logger.Error(err, "failed to request review")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
