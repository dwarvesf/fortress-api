# TASK-006: Update Swagger Documentation

## Priority
P3 - Documentation

## Estimated Effort
5 minutes

## Description
Generate Swagger documentation to include the new endpoint in the API specification.

## Dependencies
- TASK-001 (Request model)
- TASK-002 (Controller method)
- TASK-003 (Handler interface)
- TASK-004 (Handler implementation)
- TASK-005 (Route registration)

## Implementation Details

### Swagger Generation Command

The Swagger annotations are already included in TASK-004 (handler implementation). This task simply involves regenerating the Swagger documentation.

### Command to Run

```bash
make gen-swagger
```

This command will:
1. Scan all handler files for Swagger annotations
2. Parse the godoc comments
3. Generate `docs/swagger.json` and `docs/swagger.yaml`
4. Update the Swagger UI at `/swagger/index.html`

### Expected Swagger Output

The generated documentation should include:

**Endpoint:**
```yaml
paths:
  /api/v1/invoices/generate-splits:
    post:
      summary: Generate invoice splits by legacy number
      description: Generate invoice splits by querying Notion with legacy number and enqueuing worker job
      operationId: generateInvoiceSplits
      tags:
        - Invoice
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GenerateSplitsRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GenerateSplitsResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '404':
          description: Not Found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Internal Server Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

**Request Schema:**
```yaml
components:
  schemas:
    GenerateSplitsRequest:
      type: object
      required:
        - legacy_number
      properties:
        legacy_number:
          type: string
          example: "INV-2024-001"
```

**Response Schema:**
```yaml
components:
  schemas:
    GenerateSplitsResponse:
      type: object
      properties:
        legacy_number:
          type: string
          example: "INV-2024-001"
        invoice_page_id:
          type: string
          example: "2bf64b29-b84c-80e2-8cc7-000bfe534203"
        job_enqueued:
          type: boolean
          example: true
        message:
          type: string
          example: "Invoice splits generation job enqueued successfully"
```

## Verification Steps

### 1. Generate Documentation

```bash
make gen-swagger
```

### 2. Check Generated Files

```bash
# Check if files were updated
ls -lh docs/swagger.*

# View endpoint in JSON
cat docs/swagger.json | jq '.paths["/api/v1/invoices/generate-splits"]'

# Or in YAML
grep -A 20 "/generate-splits" docs/swagger.yaml
```

### 3. Verify in Swagger UI

```bash
# Start the application
make dev

# In browser, navigate to:
# http://localhost:8080/swagger/index.html

# Look for the new endpoint in the "Invoice" tag section
```

### 4. Test in Swagger UI

1. Navigate to Swagger UI
2. Find `POST /api/v1/invoices/generate-splits` under "Invoice" tag
3. Click "Try it out"
4. Enter test payload:
   ```json
   {
     "legacy_number": "INV-2024-001"
   }
   ```
5. Click "Execute"
6. Verify response matches expected structure

## Acceptance Criteria

- [ ] Swagger generation command runs without errors: `make gen-swagger`
- [ ] `docs/swagger.json` is updated with new timestamp
- [ ] `docs/swagger.yaml` is updated with new timestamp
- [ ] New endpoint appears in Swagger documentation at `/api/v1/invoices/generate-splits`
- [ ] Endpoint is tagged under "Invoice" group
- [ ] Request schema `GenerateSplitsRequest` is documented
- [ ] Response schema `GenerateSplitsResponse` is documented
- [ ] All HTTP status codes (200, 400, 404, 500) are documented
- [ ] Security requirement (BearerAuth) is included
- [ ] Swagger UI displays the endpoint correctly
- [ ] "Try it out" functionality works in Swagger UI

## Troubleshooting

### Issue: Endpoint not appearing

**Cause**: Swagger annotations might be missing or malformed in handler

**Solution**: Check handler godoc comments in TASK-004, ensure all annotations are present:
- `@Summary`
- `@Description`
- `@id`
- `@Tags`
- `@Accept`
- `@Produce`
- `@Security`
- `@Param`
- `@Success`
- `@Failure`
- `@Router`

### Issue: Schemas not generated

**Cause**: `@name` annotation missing from structs

**Solution**: Ensure structs have `@name` annotation:
```go
type GenerateSplitsRequest struct {
	// ...
} // @name GenerateSplitsRequest  <- This is required
```

### Issue: Build errors during swagger generation

**Cause**: Code compilation issues

**Solution**:
```bash
# Fix any compilation errors first
go build ./...

# Then regenerate swagger
make gen-swagger
```

## Verification Commands

```bash
# Regenerate Swagger docs
make gen-swagger

# Check generated files exist and are recent
ls -lh docs/swagger.*

# Search for new endpoint in generated JSON
cat docs/swagger.json | jq '.paths | keys | .[] | select(contains("generate-splits"))'

# Validate Swagger file
# (if swagger CLI is installed)
swagger validate docs/swagger.yaml

# Start app and check UI
make dev
# Navigate to http://localhost:8080/swagger/index.html
```

## Reference Files
- Makefile target: `gen-swagger` in project Makefile
- Swagger config: `cmd/server/main.go` (Swagger setup)
- Generated files: `docs/swagger.json`, `docs/swagger.yaml`
- Similar endpoint: Check `POST /api/v1/invoices/mark-paid` in Swagger UI for comparison
