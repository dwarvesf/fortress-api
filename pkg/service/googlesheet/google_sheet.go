package googlesheet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dwarvesf/fortress-api/pkg/config"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"strings"
)

const SpreedSheetReadOnlyScope = "https://www.googleapis.com/auth/spreadsheets.readonly\t"

type googleService struct {
	config    *oauth2.Config
	token     *oauth2.Token
	service   *sheets.Service
	appConfig *config.Config
}

// New function return Google service
func New(config *oauth2.Config, appConfig *config.Config) IService {
	return &googleService{
		config:    config,
		appConfig: appConfig,
	}
}

func (g *googleService) ensureToken(rToken string) error {
	token := &oauth2.Token{
		RefreshToken: rToken,
	}

	if !token.Valid() {
		tks := g.config.TokenSource(context.Background(), token)
		tok, err := tks.Token()
		if err != nil {
			return err
		}
		g.token = tok
	}

	return nil
}

func (g *googleService) prepareService() error {
	client := g.config.Client(context.Background(), g.token)
	service, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return errors.New("Get Sheets Service Failed " + err.Error())
	}
	g.service = service
	return nil
}

func (g *googleService) FetchSheetContent(fromIdx int) ([]DeliveryMetricRawData, error) {
	DeliveryMetricSheetID := "1KXUVyDrC9199Dp6wpT6ovIkIvZRtf455eaqwZmvTAFU"
	DeliveryMetricSheetRange := "All Data"
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return nil, err
	}

	if err := g.prepareService(); err != nil {
		return nil, err
	}
	// Fetch the content of the specified Google Sheets file
	resp, err := g.service.Spreadsheets.Values.Get(DeliveryMetricSheetID, DeliveryMetricSheetRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	// Convert the response to JSON
	jsonData, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("unable to convert sheet data to JSON: %v", err)
	}

	/// Create a struct instance to hold the data
	var sheetData SheetData

	// Unmarshal the JSON data into the struct
	err = json.Unmarshal(jsonData, &sheetData)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal sheet data: %v", err)
	}
	deliveryMetricRawData := make([]DeliveryMetricRawData, 0)

	for idx := fromIdx - 1; idx < len(sheetData.Values); idx++ {
		itm := sheetData.Values[idx]

		c0 := strings.TrimSpace(itm[0])
		c1 := strings.TrimSpace(itm[1])
		c2 := strings.TrimSpace(itm[2])
		c3 := strings.TrimSpace(itm[3])
		c4 := strings.TrimSpace(itm[4])
		c5 := strings.TrimSpace(itm[5])
		c6 := strings.TrimSpace(itm[6])

		if c4 == "" || c5 == "" || c6 == "" {
			continue
		}

		if c1 == "" {
			c1 = "0"
		}

		if c2 == "" {
			c2 = "0"
		}

		if c3 == "" {
			c3 = "0"
		}

		tmp := DeliveryMetricRawData{
			Person:        c0,
			Weight:        c1,
			Effort:        c2,
			Effectiveness: c3,
			Date:          c4,
			Project:       c5,
			Email:         c6,
		}

		deliveryMetricRawData = append(deliveryMetricRawData, tmp)
	}

	return deliveryMetricRawData, nil
}
