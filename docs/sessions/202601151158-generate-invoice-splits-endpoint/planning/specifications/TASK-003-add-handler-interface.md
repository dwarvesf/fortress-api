# TASK-003: Add Handler Interface Method

## Priority
P1 - Core Implementation

## Estimated Effort
5 minutes

## Description
Add the `GenerateSplits` method to the invoice handler interface to maintain the interface contract.

## Dependencies
- TASK-001 (Request model)
- TASK-002 (Controller method)

## File to Modify
`/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/invoice/interface.go`

## Implementation Details

### Add Interface Method

Add the method signature to the `IHandler` interface after the existing `MarkPaid` method:

**Current interface (lines 5-13):**
```go
type IHandler interface {
	GetTemplate(c *gin.Context)
	List(c *gin.Context)
	Send(c *gin.Context)
	UpdateStatus(c *gin.Context)
	CalculateCommissions(c *gin.Context)
	GenerateContractorInvoice(c *gin.Context)
	MarkPaid(c *gin.Context)
}
```

**Updated interface:**
```go
type IHandler interface {
	GetTemplate(c *gin.Context)
	List(c *gin.Context)
	Send(c *gin.Context)
	UpdateStatus(c *gin.Context)
	CalculateCommissions(c *gin.Context)
	GenerateContractorInvoice(c *gin.Context)
	MarkPaid(c *gin.Context)
	GenerateSplits(c *gin.Context)  // Add this line
}
```

### Notes on Implementation

1. **Method Naming**: Use `GenerateSplits` to match the domain terminology
2. **Signature**: Follow existing pattern with `*gin.Context` parameter
3. **Position**: Add at the end of the interface to maintain alphabetical-ish ordering
4. **No Return Values**: Handler methods use Gin context for responses

## Acceptance Criteria

- [ ] `GenerateSplits(c *gin.Context)` method is added to `IHandler` interface
- [ ] Method signature follows existing patterns
- [ ] File compiles without errors: `go build ./pkg/handler/invoice/...`
- [ ] Interface is still properly implemented by the handler struct (verified by compiler)

## Verification Commands

```bash
# Verify compilation
go build ./pkg/handler/invoice/...

# Check interface compliance (will fail if handler doesn't implement it)
go test ./pkg/handler/invoice/... -run TestInterfaceCompliance || echo "Interface check passed"
```

## Notes

This is a simple but critical task. The Go compiler will enforce that the handler struct implements this method once we add it in TASK-004. If the handler implementation is missing, compilation will fail with a clear error message.

## Reference Files
- Current interface: `pkg/handler/invoice/interface.go`
- Similar method: `MarkPaid(c *gin.Context)` in the same interface
