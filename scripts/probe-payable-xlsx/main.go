package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/googledrive"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/vault"
)

var invoiceIDPattern = regexp.MustCompile(`INVC-\d{6}-[A-Z0-9]+-[A-Z0-9]+`)

type workbook struct {
	Sheets []workbookSheet `xml:"sheets>sheet"`
}

type workbookSheet struct {
	Name string `xml:"name,attr"`
	RID  string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
}

type workbookRels struct {
	Relationships []workbookRel `xml:"Relationship"`
}

type workbookRel struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type worksheet struct {
	Rows []sheetRow `xml:"sheetData>row"`
}

type sheetRow struct {
	Index int         `xml:"r,attr"`
	Cells []sheetCell `xml:"c"`
}

type sheetCell struct {
	Ref       string      `xml:"r,attr"`
	Type      string      `xml:"t,attr"`
	Value     string      `xml:"v"`
	InlineStr *inlineCell `xml:"is"`
}

type inlineCell struct {
	Text string `xml:"t"`
}

type sharedStrings struct {
	Items []sharedStringItem `xml:"si"`
}

type sharedStringItem struct {
	Text string    `xml:"t"`
	Runs []textRun `xml:"r"`
}

type textRun struct {
	Text string `xml:"t"`
}

type sheetAnalysis struct {
	Name           string
	ColumnQValues  []string
	InvoiceIDs     []string
	MatchedRows    []string
	NonEmptyQCount int
}

type invoiceLookupResult struct {
	InvoiceID string
	PageID    string
	Status    string
	Found     bool
}

func main() {
	fileName := flag.String("name", "2026_03_ach_extra_pO.xlsx", "XLSX file name to find in Google Drive")
	flag.Parse()

	cfg := config.LoadConfig(config.DefaultConfigLoaders())
	log := logger.NewLogrusLogger("debug")

	v, err := vault.New(cfg)
	if err != nil {
		log.Warnf("vault unavailable, continuing with local config: %v", err)
	} else if v != nil {
		cfg = config.Generate(v)
	}

	l := log.Fields(logger.Fields{
		"script":    "probe-payable-xlsx",
		"file_name": *fileName,
	})

	if cfg.Google.AccountingGoogleRefreshToken == "" {
		l.Fatal(nil, "ACCOUNTING_GOOGLE_REFRESH_TOKEN is empty")
	}

	ctx := context.Background()
	driveSvc, err := newDriveService(ctx, cfg)
	if err != nil {
		l.Fatal(err, "failed to create Google Drive service")
	}

	l.Debug("searching Google Drive for target xlsx file")
	files, err := findFilesByName(ctx, driveSvc, *fileName)
	if err != nil {
		l.Fatal(err, "failed to search file by name")
	}

	if len(files) == 0 {
		fragments := buildSearchFragments(*fileName)
		candidates, candidateErr := findFilesByFragments(ctx, driveSvc, fragments)
		if candidateErr != nil {
			l.Fatal(candidateErr, "target file not found and fallback candidate search failed")
		}

		fmt.Printf("FILE_NAME=%s\n", *fileName)
		fmt.Println("EXACT_MATCH=false")
		fmt.Printf("FALLBACK_CANDIDATES=%d\n", len(candidates))
		for _, candidate := range candidates {
			fmt.Printf("- %s | %s | %s\n", candidate.Name, candidate.Id, candidate.MimeType)
		}

		l.Fatal(nil, "target file not found in accessible Google Drive scope")
	}

	for _, f := range files {
		l.Fields(logger.Fields{
			"script":    "probe-payable-xlsx",
			"file_name": *fileName,
			"file_id":   f.Id,
			"mime_type": f.MimeType,
			"size":      f.Size,
		}).Debug("matched candidate file")
	}

	target := files[0]
	l = l.Fields(logger.Fields{
		"script":    "probe-payable-xlsx",
		"file_name": *fileName,
		"file_id":   target.Id,
	})

	if target.MimeType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		l.Fatalf(nil, "target file is not an xlsx blob, mimeType=%s", target.MimeType)
	}

	l.Debug("downloading xlsx file bytes from Google Drive")
	body, err := downloadFile(ctx, driveSvc, target.Id)
	if err != nil {
		l.Fatal(err, "failed to download xlsx file")
	}
	l.AddField("downloaded_bytes", len(body)).Debug("download complete")

	analyses, err := analyzeWorkbook(body)
	if err != nil {
		l.Fatal(err, "failed to analyze xlsx workbook")
	}

	invoiceIDs := collectInvoiceIDs(analyses)
	lookupResults, err := lookupPayablesByInvoiceID(ctx, cfg, log, invoiceIDs)
	if err != nil {
		l.Fatal(err, "failed to verify invoice IDs against contractor payables")
	}

	printReport(target, analyses, lookupResults)
}

func newDriveService(ctx context.Context, cfg *config.Config) (*drive.Service, error) {
	driveConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{googledrive.FullDriveAccessScope},
	}

	tokenSource := driveConfig.TokenSource(ctx, &oauth2.Token{
		RefreshToken: cfg.Google.AccountingGoogleRefreshToken,
	})

	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("refresh access token: %w", err)
	}

	client := driveConfig.Client(ctx, token)
	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("create drive service: %w", err)
	}

	return service, nil
}

func findFilesByName(ctx context.Context, service *drive.Service, fileName string) ([]*drive.File, error) {
	result, err := service.Files.List().
		Context(ctx).
		Q(fmt.Sprintf("name='%s' and trashed=false", escapeDriveQueryValue(fileName))).
		Fields("files(id,name,mimeType,size,parents,driveId)").
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		PageSize(10).
		Do()
	if err != nil {
		return nil, err
	}

	return result.Files, nil
}

func findFilesByFragments(ctx context.Context, service *drive.Service, fragments []string) ([]*drive.File, error) {
	seen := make(map[string]*drive.File)

	for _, fragment := range fragments {
		if fragment == "" {
			continue
		}

		result, err := service.Files.List().
			Context(ctx).
			Q(fmt.Sprintf("name contains '%s' and trashed=false", escapeDriveQueryValue(fragment))).
			Fields("files(id,name,mimeType,size,parents,driveId)").
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			PageSize(20).
			Do()
		if err != nil {
			return nil, err
		}

		for _, file := range result.Files {
			seen[file.Id] = file
		}
	}

	files := make([]*drive.File, 0, len(seen))
	for _, file := range seen {
		files = append(files, file)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files, nil
}

func downloadFile(ctx context.Context, service *drive.Service, fileID string) ([]byte, error) {
	resp, err := service.Files.Get(fileID).
		Context(ctx).
		SupportsAllDrives(true).
		Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func analyzeWorkbook(data []byte) ([]sheetAnalysis, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open xlsx zip: %w", err)
	}

	files := make(map[string]*zip.File, len(reader.File))
	for _, f := range reader.File {
		files[f.Name] = f
	}

	shared, err := readSharedStrings(files)
	if err != nil {
		return nil, err
	}

	sheetTargets, err := resolveSheetTargets(files)
	if err != nil {
		return nil, err
	}

	analyses := make([]sheetAnalysis, 0, len(sheetTargets))
	for _, target := range sheetTargets {
		analysis, err := analyzeSheet(files, shared, target.Name, target.Path)
		if err != nil {
			return nil, err
		}
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

type sheetTarget struct {
	Name string
	Path string
}

func resolveSheetTargets(files map[string]*zip.File) ([]sheetTarget, error) {
	workbookData, err := readZipFile(files, "xl/workbook.xml")
	if err != nil {
		return nil, err
	}

	relsData, err := readZipFile(files, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, err
	}

	var wb workbook
	if err := xml.Unmarshal(workbookData, &wb); err != nil {
		return nil, fmt.Errorf("parse workbook.xml: %w", err)
	}

	var rels workbookRels
	if err := xml.Unmarshal(relsData, &rels); err != nil {
		return nil, fmt.Errorf("parse workbook relations: %w", err)
	}

	relMap := make(map[string]string, len(rels.Relationships))
	for _, rel := range rels.Relationships {
		relMap[rel.ID] = normalizeWorkbookTarget(rel.Target)
	}

	targets := make([]sheetTarget, 0, len(wb.Sheets))
	for _, sheet := range wb.Sheets {
		path, ok := relMap[sheet.RID]
		if !ok {
			return nil, fmt.Errorf("sheet relationship not found: %s", sheet.RID)
		}
		targets = append(targets, sheetTarget{Name: sheet.Name, Path: path})
	}

	return targets, nil
}

func readSharedStrings(files map[string]*zip.File) ([]string, error) {
	sharedFile, ok := files["xl/sharedStrings.xml"]
	if !ok {
		return nil, nil
	}

	rc, err := sharedFile.Open()
	if err != nil {
		return nil, fmt.Errorf("open sharedStrings.xml: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read sharedStrings.xml: %w", err)
	}

	var ss sharedStrings
	if err := xml.Unmarshal(data, &ss); err != nil {
		return nil, fmt.Errorf("parse sharedStrings.xml: %w", err)
	}

	values := make([]string, 0, len(ss.Items))
	for _, item := range ss.Items {
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

func analyzeSheet(files map[string]*zip.File, shared []string, sheetName, path string) (sheetAnalysis, error) {
	data, err := readZipFile(files, path)
	if err != nil {
		return sheetAnalysis{}, err
	}

	var ws worksheet
	if err := xml.Unmarshal(data, &ws); err != nil {
		return sheetAnalysis{}, fmt.Errorf("parse worksheet %s: %w", sheetName, err)
	}

	analysis := sheetAnalysis{Name: sheetName}
	for _, row := range ws.Rows {
		for _, cell := range row.Cells {
			if columnLetters(cell.Ref) != "Q" {
				continue
			}

			value, err := resolveCellValue(cell, shared)
			if err != nil {
				return sheetAnalysis{}, fmt.Errorf("resolve cell %s in sheet %s: %w", cell.Ref, sheetName, err)
			}

			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}

			analysis.NonEmptyQCount++
			analysis.ColumnQValues = append(analysis.ColumnQValues, trimmed)

			matches := invoiceIDPattern.FindAllString(strings.ToUpper(trimmed), -1)
			for _, match := range matches {
				analysis.InvoiceIDs = append(analysis.InvoiceIDs, match)
				analysis.MatchedRows = append(analysis.MatchedRows, fmt.Sprintf("%s=%s", cell.Ref, match))
			}
		}
	}

	analysis.InvoiceIDs = uniqueStrings(analysis.InvoiceIDs)
	analysis.MatchedRows = uniqueStrings(analysis.MatchedRows)
	return analysis, nil
}

func resolveCellValue(cell sheetCell, shared []string) (string, error) {
	switch cell.Type {
	case "s":
		idx, err := strconv.Atoi(strings.TrimSpace(cell.Value))
		if err != nil {
			return "", fmt.Errorf("parse shared string index: %w", err)
		}
		if idx < 0 || idx >= len(shared) {
			return "", fmt.Errorf("shared string index out of range: %d", idx)
		}
		return shared[idx], nil
	case "inlineStr":
		if cell.InlineStr == nil {
			return "", nil
		}
		return cell.InlineStr.Text, nil
	default:
		return cell.Value, nil
	}
}

func readZipFile(files map[string]*zip.File, path string) ([]byte, error) {
	f, ok := files[path]
	if !ok {
		return nil, fmt.Errorf("xlsx entry not found: %s", path)
	}

	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("open xlsx entry %s: %w", path, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read xlsx entry %s: %w", path, err)
	}

	return data, nil
}

func normalizeWorkbookTarget(target string) string {
	target = strings.TrimPrefix(target, "/")
	if strings.HasPrefix(target, "xl/") {
		return target
	}
	return "xl/" + strings.TrimPrefix(target, "../")
}

func columnLetters(cellRef string) string {
	var builder strings.Builder
	for _, r := range cellRef {
		if r < 'A' || r > 'Z' {
			break
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}

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

func escapeDriveQueryValue(value string) string {
	return strings.ReplaceAll(value, "'", "\\'")
}

func buildSearchFragments(fileName string) []string {
	trimmed := strings.TrimSpace(fileName)
	trimmed = strings.TrimSuffix(trimmed, ".xlsx")
	parts := strings.Split(trimmed, "_")

	fragments := []string{trimmed}
	if len(parts) >= 4 {
		fragments = append(fragments,
			strings.Join(parts[:4], "_"),
			strings.Join(parts[:3], "_"),
		)
	}
	if len(parts) >= 2 {
		fragments = append(fragments, strings.Join(parts[:2], "_"))
	}

	return uniqueStrings(fragments)
}

func printReport(file *drive.File, analyses []sheetAnalysis, lookupResults []invoiceLookupResult) {
	fmt.Printf("FILE_ID=%s\n", file.Id)
	fmt.Printf("FILE_NAME=%s\n", file.Name)
	fmt.Printf("MIME_TYPE=%s\n", file.MimeType)
	fmt.Printf("SHEETS=%d\n", len(analyses))

	totalNonEmpty := 0
	allInvoiceIDs := collectInvoiceIDs(analyses)
	for _, analysis := range analyses {
		totalNonEmpty += analysis.NonEmptyQCount
	}

	fmt.Printf("COLUMN_Q_NON_EMPTY=%d\n", totalNonEmpty)
	fmt.Printf("INVOICE_ID_MATCHES=%d\n", len(allInvoiceIDs))

	for _, analysis := range analyses {
		fmt.Printf("\n[SHEET] %s\n", analysis.Name)
		fmt.Printf("column_q_non_empty=%d\n", analysis.NonEmptyQCount)
		fmt.Printf("invoice_id_matches=%d\n", len(analysis.InvoiceIDs))

		if len(analysis.MatchedRows) > 0 {
			fmt.Println("matched_rows:")
			for _, row := range analysis.MatchedRows {
				fmt.Printf("- %s\n", row)
			}
		}

		if len(analysis.ColumnQValues) > 0 {
			fmt.Println("sample_column_q_values:")
			for _, value := range limitStrings(analysis.ColumnQValues, 10) {
				fmt.Printf("- %s\n", value)
			}
		}
	}

	if len(allInvoiceIDs) > 0 {
		fmt.Println("\nUNIQUE_INVOICE_IDS:")
		for _, invoiceID := range allInvoiceIDs {
			fmt.Printf("- %s\n", invoiceID)
		}
	}

	if len(lookupResults) > 0 {
		foundCount := 0
		for _, result := range lookupResults {
			if result.Found {
				foundCount++
			}
		}

		fmt.Printf("\nNOTION_PAYABLE_MATCHES=%d\n", foundCount)
		fmt.Println("PAYABLE_LOOKUPS:")
		for _, result := range lookupResults {
			if result.Found {
				fmt.Printf("- %s | FOUND | %s | %s\n", result.InvoiceID, result.PageID, result.Status)
				continue
			}
			fmt.Printf("- %s | NOT_FOUND\n", result.InvoiceID)
		}
	}
}

func collectInvoiceIDs(analyses []sheetAnalysis) []string {
	allInvoiceIDs := make([]string, 0)
	for _, analysis := range analyses {
		allInvoiceIDs = append(allInvoiceIDs, analysis.InvoiceIDs...)
	}
	allInvoiceIDs = uniqueStrings(allInvoiceIDs)
	sort.Strings(allInvoiceIDs)
	return allInvoiceIDs
}

func lookupPayablesByInvoiceID(ctx context.Context, cfg *config.Config, log logger.Logger, invoiceIDs []string) ([]invoiceLookupResult, error) {
	if len(invoiceIDs) == 0 {
		return nil, nil
	}

	svc := notion.NewContractorPayablesService(cfg, log, nil)
	if svc == nil {
		return nil, errors.New("failed to initialize contractor payables service")
	}

	results := make([]invoiceLookupResult, 0, len(invoiceIDs))
	for _, invoiceID := range invoiceIDs {
		payable, err := svc.FindPayableByInvoiceIDAnyStatus(ctx, invoiceID)
		if err != nil {
			return nil, fmt.Errorf("lookup invoice ID %s: %w", invoiceID, err)
		}

		result := invoiceLookupResult{InvoiceID: invoiceID}
		if payable != nil {
			result.Found = true
			result.PageID = payable.PageID
			result.Status = payable.Status
		}

		results = append(results, result)
	}

	return results, nil
}

func limitStrings(values []string, max int) []string {
	if len(values) <= max {
		return values
	}
	return values[:max]
}

func init() {
	if err := validateStaticAssumptions(); err != nil {
		panic(err)
	}
}

func validateStaticAssumptions() error {
	if !invoiceIDPattern.MatchString("INVC-202601-QUANG-4DRE") {
		return errors.New("invoice ID regex is invalid")
	}
	return nil
}
