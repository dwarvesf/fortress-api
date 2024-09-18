package employee

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type CheckinResponse struct {
	EmployeeID      string
	IcyAmount       float64
	TransactionID   string
	TransactionHash string
}

func (r *controller) CheckIn(discordID string, t time.Time, amount float64) (*CheckinResponse, error) {
	// Get employee by discord id
	employee, err := r.store.Employee.GetByDiscordID(r.repo.DB(), discordID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	checkinDate := t.Format("2006-01-02")

	// check if record already exists
	epc, err := r.store.PhysicalCheckin.GetByEmployeeIDAndDate(r.repo.DB(), employee.ID.String(), checkinDate)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if epc.ID != 0 {
		return nil, ErrAlreadyCheckedIn
	}

	err = r.checkFullTimeRole(employee)
	if err != nil {
		return nil, err
	}

	tx, done := r.repo.NewTransaction()
	pc := &model.PhysicalCheckinTransaction{
		EmployeeID: employee.ID,
		IcyAmount:  amount,
		Date:       t,
	}
	if err := r.store.PhysicalCheckin.Save(tx.DB(), pc); err != nil {
		return nil, done(err)
	}

	// Make transaction request to mochi
	description := fmt.Sprintf("%s - Physical Checkin on %v", employee.DisplayName, checkinDate)
	references := "Physical Checkin"
	txs, err := r.service.Mochi.SendFromAccountToUser(amount, discordID, description, references)
	if err != nil {
		return nil, done(err)
	}

	if len(txs) == 0 {
		return nil, done(ErrNoTransactionFound)
	}

	pc.MochiTxID = txs[0].TransactionID
	if err := r.store.PhysicalCheckin.Save(tx.DB(), pc); err != nil {
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
