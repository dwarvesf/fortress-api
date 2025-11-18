package nocodb

import (
	"testing"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestParseAssigneeIDs_ArrayInterface(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	// Test with []interface{} containing float64
	input := []interface{}{float64(12345678), float64(23147886)}
	result, err := service.parseAssigneeIDs(input)

	assert.NoError(t, err)
	assert.Equal(t, []int{12345678, 23147886}, result)
}

func TestParseAssigneeIDs_ArrayInterfaceWithStrings(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	// Test with []interface{} containing strings
	input := []interface{}{"12345678", "23147886"}
	result, err := service.parseAssigneeIDs(input)

	assert.NoError(t, err)
	assert.Equal(t, []int{12345678, 23147886}, result)
}

func TestParseAssigneeIDs_JSONString(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	// Test with JSON string
	input := `["12345678", "23147886"]`
	result, err := service.parseAssigneeIDs(input)

	assert.NoError(t, err)
	assert.Equal(t, []int{12345678, 23147886}, result)
}

func TestParseAssigneeIDs_CommaSeparated(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	// Test with comma-separated string
	input := "12345678,23147886"
	result, err := service.parseAssigneeIDs(input)

	assert.NoError(t, err)
	assert.Equal(t, []int{12345678, 23147886}, result)
}

func TestParseAssigneeIDs_Nil(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	// Test with nil
	result, err := service.parseAssigneeIDs(nil)

	assert.NoError(t, err)
	assert.Equal(t, []int{}, result)
}

func TestParseAssigneeIDs_EmptyString(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	// Test with empty string
	result, err := service.parseAssigneeIDs("")

	assert.NoError(t, err)
	assert.Equal(t, []int{}, result)
}

func TestGetGroups_ReturnsOutGroup(t *testing.T) {
	service := &AccountingTodoService{
		client: &Service{}, // Non-nil client
		logger: logger.NewLogrusLogger("error"),
	}

	groups, err := service.GetGroups(12345, 67890)

	assert.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.Equal(t, 12345, groups[0].ID)
	assert.Equal(t, "out", groups[0].Title)
}

func TestGetGroups_NilClient(t *testing.T) {
	service := &AccountingTodoService{
		client: nil,
	}

	groups, err := service.GetGroups(12345, 67890)

	assert.Error(t, err)
	assert.Nil(t, groups)
	assert.Equal(t, "nocodb client is nil", err.Error())
}

func TestGetLists_ReturnsDefaultList(t *testing.T) {
	service := &AccountingTodoService{
		client: &Service{}, // Non-nil client
		logger: logger.NewLogrusLogger("error"),
	}

	lists, err := service.GetLists(12345, 67890)

	assert.NoError(t, err)
	assert.Len(t, lists, 1)
	assert.Equal(t, 67890, lists[0].ID)
	assert.Equal(t, "Accounting Todos", lists[0].Name)
}

func TestGetLists_NilClient(t *testing.T) {
	service := &AccountingTodoService{
		client: nil,
	}

	lists, err := service.GetLists(12345, 67890)

	assert.Error(t, err)
	assert.Nil(t, lists)
	assert.Equal(t, "nocodb client is nil", err.Error())
}

func TestTransformRecordToTodo_MissingRecordID(t *testing.T) {
	service := &AccountingTodoService{}

	record := map[string]interface{}{
		"title": "Test Todo",
	}

	todo, err := service.transformRecordToTodo(record)

	assert.Error(t, err)
	assert.Nil(t, todo)
	assert.Equal(t, "missing record ID", err.Error())
}

func TestTransformRecordToTodo_MissingTitle(t *testing.T) {
	service := &AccountingTodoService{}

	record := map[string]interface{}{
		"Id": "123",
	}

	todo, err := service.transformRecordToTodo(record)

	assert.Error(t, err)
	assert.Nil(t, todo)
	assert.Equal(t, "missing title", err.Error())
}

func TestTransformRecordToTodo_MultipleAssignees(t *testing.T) {
	t.Skip("Test requires mock store - skipping for now")
}

func TestParseTodoTitle_ValidFormat(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	description, amount, currency, err := service.parseTodoTitle("Tiền điện 12/2025 | 25826000 | VND")

	assert.NoError(t, err)
	assert.Equal(t, "Tiền điện 12/2025", description)
	assert.Equal(t, 25826000.0, amount)
	assert.Equal(t, "VND", currency)
}

func TestParseTodoTitle_WithThousandSeparators(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	description, amount, currency, err := service.parseTodoTitle("Office Rental | 24,000,000 | VND")

	assert.NoError(t, err)
	assert.Equal(t, "Office Rental", description)
	assert.Equal(t, 24000000.0, amount)
	assert.Equal(t, "VND", currency)
}

func TestParseTodoTitle_WithDotSeparators(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	description, amount, currency, err := service.parseTodoTitle("CBRE 12/2025 | 2.210.000 | VND")

	assert.NoError(t, err)
	assert.Equal(t, "CBRE 12/2025", description)
	assert.Equal(t, 2210000.0, amount)
	assert.Equal(t, "VND", currency)
}

func TestParseTodoTitle_InvalidFormat(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	_, _, _, err := service.parseTodoTitle("Invalid title without pipes")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid title format")
}

func TestParseTodoTitle_InvalidAmount(t *testing.T) {
	service := &AccountingTodoService{
		logger: logger.NewLogrusLogger("error"),
	}

	_, _, _, err := service.parseTodoTitle("Test | invalid_amount | VND")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse amount")
}
