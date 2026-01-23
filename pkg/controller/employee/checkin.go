package employee

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type CheckinResponse struct {
	EmployeeID      string
	IcyAmount       float64
	TransactionID   string
	TransactionHash string
}

func (r *controller) CheckIn(discordID string, t time.Time, amount float64) (*CheckinResponse, error) {
	l := r.logger.Fields(logger.Fields{
		"controller": "employee",
		"method":     "CheckIn",
	})

	// Get employee by discord id
	employee, err := r.store.Employee.GetByDiscordID(r.repo.DB(), discordID, true)
	if err != nil {
		l.Error(err, "failed to get employee by discord id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	checkinDate := t.Format("2006-01-02")

	// check if record already exists
	epc, err := r.store.PhysicalCheckin.GetByEmployeeIDAndDate(r.repo.DB(), employee.ID.String(), checkinDate)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(err, "failed to get physical checkin by employee id and date")
		return nil, err
	}
	if !epc.ID.IsZero() {
		return nil, ErrAlreadyCheckedIn
	}

	tx, done := r.repo.NewTransaction()
	pc := &model.PhysicalCheckinTransaction{
		ID:         model.NewUUID(),
		EmployeeID: employee.ID,
		IcyAmount:  amount,
		Date:       t,
	}
	if err := r.store.PhysicalCheckin.Save(tx.DB(), pc); err != nil {
		l.Error(err, "failed to save physical checkin")
		return nil, done(err)
	}

	// Make transaction request to mochi
	description := fmt.Sprintf("%s - Physical Checkin on %v", employee.DisplayName, checkinDate)
	references := "Physical Checkin"
	txs, err := r.service.Mochi.SendFromAccountToUser(amount, discordID, description, references)
	if err != nil {
		l.Error(err, "failed to request to mochi")
		return nil, done(err)
	}

	if len(txs) == 0 {
		return nil, done(ErrNoTransactionFound)
	}

	pc.MochiTxID = txs[0].TransactionID
	if err := r.store.PhysicalCheckin.Save(tx.DB(), pc); err != nil {
		l.Error(err, "failed to save physical checkin")
		return nil, done(err)
	}

	response := &CheckinResponse{
		EmployeeID:      employee.ID.String(),
		IcyAmount:       amount,
		TransactionID:   strconv.Itoa(int(txs[0].TransactionID)),
		TransactionHash: txs[0].RecipientID,
	}

	return response, done(nil)
}

// isEmployeeWhitelisted checks if an employee is whitelisted for check-in
func (r *controller) isEmployeeWhitelisted(employeeID model.UUID) bool {
	for _, id := range r.config.CheckIn.WhitelistedEmployeeIDs {
		if id == employeeID.String() {
			return true
		}
	}
	return false
}
