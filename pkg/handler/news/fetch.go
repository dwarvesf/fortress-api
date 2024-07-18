package news

import (
	"net/http"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
)

func (h *handler) Fetch(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "discord",
			"method":  "Fetch",
		},
	)

	platform := strings.TrimSpace(c.Query("platform"))
	if platform == "" {
		l.Info("platform is empty")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, nil, nil, "platform is empty"))
		return
	}

	topic := strings.TrimSpace(c.Query("topic"))
	if topic == "" {
		l.Info("topic is empty")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, nil, nil, "topic is empty"))
		return
	}

	var popular, emerging []model.News
	var err error
	switch platform {
	case "lobsters":
		popular, emerging, err = h.controller.News.FetchLobstersNews(c.Request.Context(), topic)
		if err != nil {
			l.Error(err, "failed to fetch lobsters news")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToFetchNewsResponse(popular, emerging), nil, nil, nil, ""))
}
