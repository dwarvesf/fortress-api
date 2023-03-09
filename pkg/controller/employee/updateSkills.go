package employee

import (
	"errors"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UpdateSkillsInput struct {
	Positions       []model.UUID
	LeadingChapters []model.UUID
	Chapters        []model.UUID
	Seniority       model.UUID
	Stacks          []model.UUID
}

func (r *controller) UpdateSkills(l logger.Logger, employeeID string, body UpdateSkillsInput) (*model.Employee, error) {
	emp, err := r.store.Employee.One(r.repo.DB(), employeeID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	// Check chapter existence
	chapters, err := r.store.Chapter.All(r.repo.DB())
	if err != nil {
		return nil, err
	}

	chapterMap := model.ToChapterMap(chapters)
	for _, sID := range body.Chapters {
		_, ok := chapterMap[sID]
		if !ok {
			l.Errorf(ErrChapterNotFound, "chapter not found with id ", sID.String())
			return nil, ErrChapterNotFound
		}
	}

	// Check seniority existence
	exist, err := r.store.Seniority.IsExist(r.repo.DB(), body.Seniority.String())
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, ErrSeniorityNotFound
	}

	// Check stack existence
	_, stacks, err := r.store.Stack.All(r.repo.DB(), "", nil)
	if err != nil {
		return nil, err
	}

	stackMap := model.ToStackMap(stacks)
	for _, sID := range body.Stacks {
		_, ok := stackMap[sID]
		if !ok {
			l.Errorf(ErrStackNotFound, "stack not found with id ", sID.String())
			return nil, ErrStackNotFound
		}
	}

	// Check position existence
	positions, err := r.store.Position.All(r.repo.DB())
	if err != nil {
		return nil, err
	}

	positionMap := model.ToPositionMap(positions)
	for _, pID := range body.Positions {
		_, ok := positionMap[pID]

		if !ok {
			l.Errorf(ErrPositionNotFound, "position not found with id ", pID.String())
			return nil, ErrPositionNotFound
		}
	}

	// Begin transaction
	tx, done := r.repo.NewTransaction()

	// Delete all exist employee positions
	if err := r.store.EmployeePosition.DeleteByEmployeeID(tx.DB(), employeeID); err != nil {
		return nil, done(err)
	}

	// Create new employee position
	for _, positionID := range body.Positions {
		_, err := r.store.EmployeePosition.Create(tx.DB(), &model.EmployeePosition{
			EmployeeID: model.MustGetUUIDFromString(employeeID),
			PositionID: positionID,
		})
		if err != nil {
			return nil, done(err)
		}
	}

	// Delete all exist employee stack
	if err := r.store.EmployeeStack.DeleteByEmployeeID(tx.DB(), employeeID); err != nil {
		return nil, done(err)
	}

	// Create new employee stack
	for _, stackID := range body.Stacks {
		_, err := r.store.EmployeeStack.Create(tx.DB(), &model.EmployeeStack{
			EmployeeID: model.MustGetUUIDFromString(employeeID),
			StackID:    stackID,
		})
		if err != nil {
			return nil, done(err)
		}
	}

	// Delete all exist employee stack
	if err := r.store.EmployeeChapter.DeleteByEmployeeID(tx.DB(), employeeID); err != nil {
		return nil, done(err)
	}

	// Create new employee stack
	for _, chapterID := range body.Chapters {
		_, err := r.store.EmployeeChapter.Create(tx.DB(), &model.EmployeeChapter{
			EmployeeID: model.MustGetUUIDFromString(employeeID),
			ChapterID:  chapterID,
		})
		if err != nil {
			return nil, done(err)
		}
	}

	// Remove all chapter lead by employee
	leadingChapters, err := r.store.Chapter.GetAllByLeadID(tx.DB(), employeeID)
	if err != nil {
		return nil, done(err)
	}

	for _, lChapter := range leadingChapters {
		if err := r.store.Chapter.UpdateChapterLead(tx.DB(), lChapter.ID.String(), nil); err != nil {
			return nil, done(err)
		}
	}

	// Create new chapter
	leader := model.MustGetUUIDFromString(employeeID)
	for _, lChapter := range body.LeadingChapters {
		if err := r.store.Chapter.UpdateChapterLead(tx.DB(), lChapter.String(), &leader); err != nil {
			return nil, done(err)
		}
	}

	// Update employee information
	emp.SeniorityID = body.Seniority

	_, err = r.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, *emp, "chapter_id", "seniority_id")
	if err != nil {
		return nil, done(err)
	}

	return emp, done(nil)
}
