package contractorpayables

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

var contractorPaymentInvoicePattern = regexp.MustCompile(`INVC-\d{6}-[A-Z0-9]+-[A-Z0-9]+`)

type workbookFile struct {
	Sheets []workbookSheet `xml:"sheets>sheet"`
}

type workbookSheet struct {
	Name string `xml:"name,attr"`
	RID  string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
}

type workbookRelationships struct {
	Relationships []workbookRelationship `xml:"Relationship"`
}

type workbookRelationship struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type worksheetFile struct {
	Rows []worksheetRow `xml:"sheetData>row"`
}

type worksheetRow struct {
	Cells []worksheetCell `xml:"c"`
}

type worksheetCell struct {
	Ref       string           `xml:"r,attr"`
	Type      string           `xml:"t,attr"`
	Value     string           `xml:"v"`
	InlineStr *worksheetInline `xml:"is"`
}

type worksheetInline struct {
	Text string `xml:"t"`
}

type workbookSharedStrings struct {
	Items []workbookSharedString `xml:"si"`
}

type workbookSharedString struct {
	Text string                    `xml:"t"`
	Runs []workbookSharedStringRun `xml:"r"`
}

type workbookSharedStringRun struct {
	Text string `xml:"t"`
}

func (c *controller) CommitPayablesByFile(ctx context.Context, fileName string, year int) (*CommitResponse, error) {
	resolvedName := normalizePaymentFileName(fileName)
	l := c.logger.Fields(logger.Fields{
		"controller": "contractorpayables",
		"method":     "CommitPayablesByFile",
		"file_name":  resolvedName,
		"year":       year,
	})

	if c.config.Invoice.ContractorPaymentDirID == "" {
		return nil, fmt.Errorf("CONTRACTOR_PAYMENT_DIR_ID is not configured")
	}

	l.AddField("parent_dir_id", c.config.Invoice.ContractorPaymentDirID).Debug("starting file-based payable commit")

	body, err := c.service.GoogleDrive.DownloadFileFromYearDir(c.config.Invoice.ContractorPaymentDirID, strconv.Itoa(year), resolvedName)
	if err != nil {
		l.Error(err, "failed to download contractor payment file from Google Drive")
		return nil, fmt.Errorf("failed to download contractor payment file: %w", err)
	}

	invoiceIDs, err := extractInvoiceIDsFromWorkbook(body)
	if err != nil {
		l.Error(err, "failed to extract invoice IDs from payment file")
		return nil, fmt.Errorf("failed to parse contractor payment file: %w", err)
	}
	if len(invoiceIDs) == 0 {
		return nil, fmt.Errorf("no invoice IDs found in column Q for file %s", resolvedName)
	}

	l.AddField("invoice_id_count", len(invoiceIDs)).Debug("extracted invoice IDs from payment file")

	toCommit := make([]payableToCommit, 0, len(invoiceIDs))
	errors := make([]CommitError, 0)
	seenPayables := make(map[string]struct{})

	for _, invoiceID := range invoiceIDs {
		payable, lookupErr := c.service.Notion.ContractorPayables.FindPendingPayableByInvoiceID(ctx, invoiceID)
		if lookupErr != nil {
			errors = append(errors, CommitError{InvoiceID: invoiceID, Error: lookupErr.Error()})
			continue
		}
		if payable == nil {
			errors = append(errors, CommitError{InvoiceID: invoiceID, Error: "pending payable not found"})
			continue
		}
		if _, exists := seenPayables[payable.PageID]; exists {
			continue
		}

		seenPayables[payable.PageID] = struct{}{}
		toCommit = append(toCommit, payableToCommit{
			PageID:            payable.PageID,
			ContractorPageID:  payable.ContractorPageID,
			PayoutItemPageIDs: payable.PayoutItemPageIDs,
		})
	}

	if len(toCommit) == 0 {
		return nil, fmt.Errorf("no pending payables found for file %s", resolvedName)
	}

	successCount := 0
	failCount := len(errors)
	var mu sync.Mutex
	g := new(errgroup.Group)

	for _, payable := range toCommit {
		g.Go(func() error {
			if commitErr := c.commitSinglePayable(ctx, payable); commitErr != nil {
				mu.Lock()
				failCount++
				errors = append(errors, CommitError{PayableID: payable.PageID, Error: commitErr.Error()})
				mu.Unlock()
				return nil
			}

			mu.Lock()
			successCount++
			mu.Unlock()
			return nil
		})
	}

	_ = g.Wait()

	result := &CommitResponse{
		Mode:     "file",
		FileName: resolvedName,
		Year:     year,
		Updated:  successCount,
		Failed:   failCount,
		Errors:   errors,
	}

	l.AddField("updated", result.Updated).AddField("failed", result.Failed).Debug("completed file-based payable commit")
	return result, nil
}

func normalizePaymentFileName(fileName string) string {
	trimmed := strings.TrimSpace(filepath.Base(fileName))
	if !strings.HasSuffix(strings.ToLower(trimmed), ".xlsx") {
		trimmed += ".xlsx"
	}
	return trimmed
}

func extractInvoiceIDsFromWorkbook(data []byte) ([]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open xlsx zip: %w", err)
	}

	files := make(map[string]*zip.File, len(reader.File))
	for _, file := range reader.File {
		files[file.Name] = file
	}

	sharedStrings, err := readWorkbookSharedStrings(files)
	if err != nil {
		return nil, err
	}

	sheetTargets, err := resolveWorkbookSheetTargets(files)
	if err != nil {
		return nil, err
	}

	invoiceIDs := make([]string, 0)
	for _, target := range sheetTargets {
		values, extractErr := extractColumnQValues(files, sharedStrings, target.path)
		if extractErr != nil {
			return nil, extractErr
		}

		for _, value := range values {
			matches := contractorPaymentInvoicePattern.FindAllString(strings.ToUpper(strings.TrimSpace(value)), -1)
			invoiceIDs = append(invoiceIDs, matches...)
		}
	}

	invoiceIDs = uniqueInvoiceIDs(invoiceIDs)
	sort.Strings(invoiceIDs)
	return invoiceIDs, nil
}

type workbookSheetTarget struct {
	path string
}

func resolveWorkbookSheetTargets(files map[string]*zip.File) ([]workbookSheetTarget, error) {
	workbookData, err := readWorkbookEntry(files, "xl/workbook.xml")
	if err != nil {
		return nil, err
	}
	relationshipsData, err := readWorkbookEntry(files, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, err
	}

	var workbook workbookFile
	if err := xml.Unmarshal(workbookData, &workbook); err != nil {
		return nil, fmt.Errorf("parse workbook.xml: %w", err)
	}
	var relationships workbookRelationships
	if err := xml.Unmarshal(relationshipsData, &relationships); err != nil {
		return nil, fmt.Errorf("parse workbook relationships: %w", err)
	}

	relMap := make(map[string]string, len(relationships.Relationships))
	for _, relationship := range relationships.Relationships {
		relMap[relationship.ID] = normalizeWorkbookTarget(relationship.Target)
	}

	targets := make([]workbookSheetTarget, 0, len(workbook.Sheets))
	for _, sheet := range workbook.Sheets {
		path, ok := relMap[sheet.RID]
		if !ok {
			return nil, fmt.Errorf("sheet relationship not found: %s", sheet.RID)
		}
		targets = append(targets, workbookSheetTarget{path: path})
	}

	return targets, nil
}

func readWorkbookSharedStrings(files map[string]*zip.File) ([]string, error) {
	sharedStringsFile, ok := files["xl/sharedStrings.xml"]
	if !ok {
		return nil, nil
	}

	rc, err := sharedStringsFile.Open()
	if err != nil {
		return nil, fmt.Errorf("open shared strings: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read shared strings: %w", err)
	}

	var shared workbookSharedStrings
	if err := xml.Unmarshal(data, &shared); err != nil {
		return nil, fmt.Errorf("parse shared strings: %w", err)
	}

	values := make([]string, 0, len(shared.Items))
	for _, item := range shared.Items {
		if item.Text != "" {
			values = append(values, item.Text)
			continue
		}

		var builder strings.Builder
		for _, run := range item.Runs {
			builder.WriteString(run.Text)
		}
		values = append(values, builder.String())
	}

	return values, nil
}

func extractColumnQValues(files map[string]*zip.File, sharedStrings []string, sheetPath string) ([]string, error) {
	data, err := readWorkbookEntry(files, sheetPath)
	if err != nil {
		return nil, err
	}

	var worksheet worksheetFile
	if err := xml.Unmarshal(data, &worksheet); err != nil {
		return nil, fmt.Errorf("parse worksheet %s: %w", sheetPath, err)
	}

	values := make([]string, 0)
	for _, row := range worksheet.Rows {
		for _, cell := range row.Cells {
			if worksheetColumn(cell.Ref) != "Q" {
				continue
			}

			value, resolveErr := resolveWorksheetCellValue(cell, sharedStrings)
			if resolveErr != nil {
				return nil, fmt.Errorf("resolve worksheet cell %s: %w", cell.Ref, resolveErr)
			}
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			values = append(values, trimmed)
		}
	}

	return values, nil
}

func resolveWorksheetCellValue(cell worksheetCell, sharedStrings []string) (string, error) {
	switch cell.Type {
	case "s":
		index, err := strconv.Atoi(strings.TrimSpace(cell.Value))
		if err != nil {
			return "", fmt.Errorf("parse shared string index: %w", err)
		}
		if index < 0 || index >= len(sharedStrings) {
			return "", fmt.Errorf("shared string index out of range: %d", index)
		}
		return sharedStrings[index], nil
	case "inlineStr":
		if cell.InlineStr == nil {
			return "", nil
		}
		return cell.InlineStr.Text, nil
	default:
		return cell.Value, nil
	}
}

func readWorkbookEntry(files map[string]*zip.File, path string) ([]byte, error) {
	file, ok := files[path]
	if !ok {
		return nil, fmt.Errorf("xlsx entry not found: %s", path)
	}

	rc, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open xlsx entry %s: %w", path, err)
	}
	defer rc.Close()

	body, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read xlsx entry %s: %w", path, err)
	}

	return body, nil
}

func normalizeWorkbookTarget(target string) string {
	target = strings.TrimPrefix(target, "/")
	if strings.HasPrefix(target, "xl/") {
		return target
	}
	return "xl/" + strings.TrimPrefix(target, "../")
}

func worksheetColumn(cellRef string) string {
	var builder strings.Builder
	for _, r := range cellRef {
		if r < 'A' || r > 'Z' {
			break
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func uniqueInvoiceIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
