# TASK-001: Add Request Model for Generate Splits

## Priority
P0 - Foundation (must be completed first)

## Estimated Effort
15 minutes

## Description
Create the request model struct for the generate splits endpoint in the existing request package.

## Dependencies
None - This is a foundational task

## File to Modify
`/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/invoice/request/request.go`

## Implementation Details

### Add Request Struct

Add the following struct after the existing `MarkPaidRequest` struct (around line 27):

```go
// GenerateSplitsRequest is the request body for generating invoice splits by legacy number
type GenerateSplitsRequest struct {
	LegacyNumber string `json:"legacy_number" binding:"required"`
} // @name GenerateSplitsRequest
```

### Add Validation Method

Add validation method after the struct:

```go
// Validate checks if the request is valid
func (r *GenerateSplitsRequest) Validate() error {
	if strings.TrimSpace(r.LegacyNumber) == "" {
		return errs.ErrInvalidInvoiceNumber
	}
	return nil
}
```

### Notes on Implementation

1. **Naming Convention**: Follow existing pattern with `@name` annotation for Swagger documentation
2. **Validation Tag**: Use `binding:"required"` for Gin's automatic validation
3. **Error Handling**: Reuse existing error constants from `pkg/handler/invoice/errs` package
4. **Field Naming**: Use `legacy_number` in JSON (snake_case) and `LegacyNumber` in struct (PascalCase)

### Error Constant Check

If `errs.ErrInvalidInvoiceNumber` doesn't exist, add it to `pkg/handler/invoice/errs/errs.go`:

```go
var ErrInvalidInvoiceNumber = errors.New("invalid invoice number")
```

## Testing Requirements

### Unit Test Cases

Create test file: `pkg/handler/invoice/request/request_test.go` (or add to existing test file)

```go
func TestGenerateSplitsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request GenerateSplitsRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: GenerateSplitsRequest{
				LegacyNumber: "INV-2024-001",
			},
			wantErr: false,
		},
		{
			name: "empty legacy number",
			request: GenerateSplitsRequest{
				LegacyNumber: "",
			},
			wantErr: true,
		},
		{
			name: "whitespace only legacy number",
			request: GenerateSplitsRequest{
				LegacyNumber: "   ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

## Acceptance Criteria

- [ ] `GenerateSplitsRequest` struct is added with correct JSON tags
- [ ] `Validate()` method is implemented and handles empty/whitespace strings
- [ ] Swagger annotation `@name` is present for API documentation
- [ ] Error constant `ErrInvalidInvoiceNumber` exists and is used
- [ ] Unit tests cover valid and invalid cases
- [ ] All tests pass: `go test ./pkg/handler/invoice/request/...`
- [ ] Code follows existing patterns in the file

## Verification Commands

```bash
# Run tests
go test ./pkg/handler/invoice/request/... -v

# Verify struct is parseable
go build ./pkg/handler/invoice/request/...
```

## Reference Files
- Similar request struct: `MarkPaidRequest` (line 24-27 in same file)
- Error constants: `pkg/handler/invoice/errs/errs.go`
- Validation pattern: `UpdateStatusRequest.Validate()` (line 29-35 in same file)
