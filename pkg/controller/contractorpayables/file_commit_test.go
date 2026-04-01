package contractorpayables

import (
	"archive/zip"
	"bytes"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePaymentFileName(t *testing.T) {
	require.Equal(t, "sample.xlsx", normalizePaymentFileName("sample"))
	require.Equal(t, "sample.xlsx", normalizePaymentFileName("sample.xlsx"))
	require.Equal(t, "sample.xlsx", normalizePaymentFileName("../sample.xlsx"))
}

func TestExtractInvoiceIDsFromWorkbook(t *testing.T) {
	workbook := buildTestWorkbook(t, []string{"Description (optional)", "INVC-202603-THANHPD-K5KW", "INVC-202603-HUYTQ-ID6I", "ignore-me"})

	invoiceIDs, err := extractInvoiceIDsFromWorkbook(workbook)
	require.NoError(t, err)
	require.Equal(t, []string{"INVC-202603-HUYTQ-ID6I", "INVC-202603-THANHPD-K5KW"}, invoiceIDs)
}

func buildTestWorkbook(t *testing.T, columnQValues []string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	writeZipEntry(t, zw, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Sheet1" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`)

	writeZipEntry(t, zw, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`)

	var sheet bytes.Buffer
	sheet.WriteString(`<?xml version="1.0" encoding="UTF-8"?><worksheet><sheetData>`)
	for idx, value := range columnQValues {
		rowNumber := strconv.Itoa(idx + 1)
		sheet.WriteString(`<row r="`)
		sheet.WriteString(rowNumber)
		sheet.WriteString(`"><c r="Q`)
		sheet.WriteString(rowNumber)
		sheet.WriteString(`" t="inlineStr"><is><t>`)
		sheet.WriteString(value)
		sheet.WriteString(`</t></is></c></row>`)
	}
	sheet.WriteString(`</sheetData></worksheet>`)

	writeZipEntry(t, zw, "xl/worksheets/sheet1.xml", sheet.String())
	require.NoError(t, zw.Close())

	return buf.Bytes()
}

func writeZipEntry(t *testing.T, zw *zip.Writer, name, content string) {
	t.Helper()
	writer, err := zw.Create(name)
	require.NoError(t, err)
	_, err = writer.Write([]byte(content))
	require.NoError(t, err)
}
