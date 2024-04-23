package communitynft

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	ctrl "github.com/dwarvesf/fortress-api/pkg/controller/communitynft"
	"github.com/dwarvesf/fortress-api/pkg/handler/communitynft/errs"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
	controller *controller.Controller
}

// New returns a handler
func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:      store,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
		controller: controller,
	}
}

// GetNftMetadata godoc
// @Summary Get metadata of a nft
// @Description Get metadata of a nft
// @id getNftMetadata
// @Tags CommunityNft
// @Accept  json
// @Produce  json
// @Param id path string true "NFT ID"
// @Success 200 {object} GetNftMetadataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /community-nfts/{id} [get]
func (h *handler) GetNftMetadata(c *gin.Context) {
	tokenIdStr := c.Param("id")
	if tokenIdStr == "" {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidTokenID, nil, ""))
		return
	}
	tid, err := strconv.Atoi(tokenIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidTokenID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "client",
		"method":  "Detail",
	})

	nftMetadata, err := h.controller.CommunityNft.GetNftMetadata(tid)
	if err == ctrl.ErrTokenNotFound {
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrTokenNotFound, nil, ""))
		return
	}
	if err != nil {
		l.Error(err, "failed to get nft metadata")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToNftMetadata(nftMetadata), nil, nil, nil, ""))
}
