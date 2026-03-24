package invoice

import (
	"bytes"
	"math"
	"path/filepath"
	"runtime"
	"testing"
	"text/template"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/stretchr/testify/assert"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func TestPDFTemplate_FloatingPointRounding(t *testing.T) {
	// Setup template functions similar to the ones in controller
	currencyCode := "USD"
	pound := money.New(1, currencyCode)

	funcMap := template.FuncMap{
		"formatMoney": func(m float64) string {
			tmpValue := math.Round(m * math.Pow(10, float64(pound.Currency().Fraction)))
			return pound.Multiply(int64(tmpValue)).Display()
		},
		"round": func(v float64) float64 {
			return math.Round(v*100) / 100
		},
		"subtract": func(a, b float64) float64 {
			return a - b
		},
		"haveDiscountColumn": func() bool {
			return false
		},
	}

	// Define a simplified version of the template for testing the discount row
	tmplStr := `
        <tr class="subtotal">
          <td class="font-bold text-right">Subtotal</td>
          <td>{{formatMoney .Invoice.SubTotal}}</td>
        </tr>
        {{if gt (round .Invoice.Discount) 0.0}}
        <tr class="discount-explicit">
          <td class="font-bold text-right">Discount {{.Invoice.DiscountType}}</td>
          <td>-{{formatMoney .Invoice.Discount}}</td>
        </tr>
        {{else if lt (round .Invoice.Total) (round .Invoice.SubTotal)}}
        <tr class="discount-implicit">
          <td class="font-bold text-right">Discount</td>
          <td>-{{formatMoney (subtract .Invoice.SubTotal .Invoice.Total)}}</td>
        </tr>
        {{end}}
        <tr class="total">
          <td class="font-bold text-right">Total</td>
          <td class="total-price font-bold">{{formatMoney .Invoice.Total}}</td>
        </tr>`

	tmpl, err := template.New("test").Funcs(funcMap).Parse(tmplStr)
	assert.NoError(t, err)

	tests := []struct {
		name             string
		subtotal         float64
		total            float64
		discount         float64
		expectDiscountRow bool
	}{
		{
			name:             "Perfect match - no discount",
			subtotal:         15636.24,
			total:            15636.24,
			discount:         0.0,
			expectDiscountRow: false,
		},
		{
			name:             "Floating point precision discrepancy - should NOT show discount",
			subtotal:         15636.240000000001,
			total:            15636.24,
			discount:         0.0,
			expectDiscountRow: false,
		},
		{
			name:             "Floating point precision discrepancy 2 - should NOT show discount",
			subtotal:         15636.24,
			total:            15636.239999999999,
			discount:         0.0,
			expectDiscountRow: false,
		},
		{
			name:             "Real discount - should show discount",
			subtotal:         15636.24,
			total:            15600.00,
			discount:         0.0,
			expectDiscountRow: true,
		},
		{
			name:             "Explicit discount - should show discount",
			subtotal:         15636.24,
			total:            15600.00,
			discount:         36.24,
			expectDiscountRow: true,
		},
		{
			name:             "Tiny explicit discount (below rounding) - should NOT show discount",
			subtotal:         15636.24,
			total:            15636.24,
			discount:         0.0000001,
			expectDiscountRow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := struct {
				Invoice *model.Invoice
			}{
				Invoice: &model.Invoice{
					SubTotal: tt.subtotal,
					Total:    tt.total,
					Discount: tt.discount,
				},
			}

			var buf bytes.Buffer
			err := tmpl.Execute(&buf, data)
			assert.NoError(t, err)

			output := buf.String()
			if tt.expectDiscountRow {
				assert.Contains(t, output, "Discount", "Expected discount row in output for %s", tt.name)
			} else {
				assert.NotContains(t, output, "Discount", "Did not expect discount row in output for %s", tt.name)
			}

			// Verify money formatting always shows .24 for the 15636.24 case
			if math.Abs(tt.subtotal-15636.24) < 0.001 {
				assert.Contains(t, output, "15,636.24")
			}
		})
	}
}

func TestFormatMoney_Rounding(t *testing.T) {
	currencyCode := "USD"
	pound := money.New(1, currencyCode)
	
	formatMoney := func(m float64) string {
		tmpValue := math.Round(m * math.Pow(10, float64(pound.Currency().Fraction)))
		return pound.Multiply(int64(tmpValue)).Display()
	}

	assert.Equal(t, "$15,636.24", formatMoney(15636.24))
	assert.Equal(t, "$15,636.24", formatMoney(15636.240000000001))
	assert.Equal(t, "$15,636.24", formatMoney(15636.239999999999))
	assert.Equal(t, "$15,636.24", formatMoney(15636.244))
	assert.Equal(t, "$15,636.25", formatMoney(15636.245))
}

func TestActualTemplateFile_Loading(t *testing.T) {
	// This test ensures the actual template file still parses correctly with our changes
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	projectRoot := filepath.Join(dir, "../../..")
	templatePath := filepath.Join(projectRoot, "pkg/templates/invoice.html")

	funcMap := template.FuncMap{
		"toString": func(month int) string { return "" },
		"formatDate": func(t *time.Time) string { return "" },
		"lastDayOfMonth": func() string { return "" },
		"formatMoney": func(money float64) string { return "" },
		"haveDescription": func(description string) bool { return false },
		"haveNote": func(note string) bool { return false },
		"haveDiscountColumn": func() bool { return false },
		"float": func(n float64) string { return "" },
		"round": func(v float64) float64 { return 0 },
		"debugValue": func(label string, value interface{}) string { return "" },
		"formatDiscount": func(discountValue float64, discountType string) string { return "" },
		"subtract": func(a, b float64) float64 { return 0 },
	}

	tmpl, err := template.New("invoice.html").Funcs(funcMap).ParseFiles(templatePath)
	assert.NoError(t, err, "Actual template file should be parsable")
	assert.NotNil(t, tmpl)
}
