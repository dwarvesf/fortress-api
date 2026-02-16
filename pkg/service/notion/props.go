package notion

import (
	"strings"
	"time"

	nt "github.com/dstotijn/go-notion"
)

// ExtractFirstRelationID extracts the first relation page ID from a relation property.
func ExtractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.Relation) == 0 {
		return ""
	}
	return prop.Relation[0].ID
}

// ExtractAllRelationIDs extracts all relation page IDs from a relation property.
func ExtractAllRelationIDs(props nt.DatabasePageProperties, propName string) []string {
	prop, ok := props[propName]
	if !ok || len(prop.Relation) == 0 {
		return nil
	}
	ids := make([]string, len(prop.Relation))
	for i, rel := range prop.Relation {
		ids[i] = rel.ID
	}
	return ids
}

// ExtractNumber extracts a number property value.
func ExtractNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Number == nil {
		return 0
	}
	return *prop.Number
}

// ExtractSelect extracts a select property value name.
func ExtractSelect(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Select == nil {
		return ""
	}
	return prop.Select.Name
}

// ExtractStatus extracts a status property value name.
func ExtractStatus(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Status == nil {
		return ""
	}
	return prop.Status.Name
}

// ExtractCheckbox extracts a checkbox property value.
func ExtractCheckbox(props nt.DatabasePageProperties, propName string) bool {
	prop, ok := props[propName]
	if !ok || prop.Checkbox == nil {
		return false
	}
	return *prop.Checkbox
}

// ExtractMultiSelectNames extracts all multi-select option names.
func ExtractMultiSelectNames(props nt.DatabasePageProperties, propName string) []string {
	prop, ok := props[propName]
	if !ok || len(prop.MultiSelect) == 0 {
		return nil
	}
	names := make([]string, len(prop.MultiSelect))
	for i, opt := range prop.MultiSelect {
		names[i] = opt.Name
	}
	return names
}

// ExtractTitle extracts a title property value, joining all title parts.
func ExtractTitle(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.Title) == 0 {
		return ""
	}
	var b strings.Builder
	for _, rt := range prop.Title {
		b.WriteString(rt.PlainText)
	}
	return b.String()
}

// ExtractRichText extracts a rich text property value, joining all parts.
func ExtractRichText(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.RichText) == 0 {
		return ""
	}
	var b strings.Builder
	for _, rt := range prop.RichText {
		b.WriteString(rt.PlainText)
	}
	return b.String()
}

// ExtractRichTextFirst extracts only the first rich text item's plain text.
// Used by BankAccount and ContractorPayables where only the first item matters.
func ExtractRichTextFirst(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || len(prop.RichText) == 0 {
		return ""
	}
	return prop.RichText[0].PlainText
}

// ExtractDate extracts a date property value as *time.Time.
func ExtractDate(props nt.DatabasePageProperties, propName string) *time.Time {
	prop, ok := props[propName]
	if !ok || prop.Date == nil {
		return nil
	}
	t := prop.Date.Start.Time
	if t.IsZero() {
		return nil
	}
	return &t
}

// ExtractDateString extracts a date property value formatted as "2006-01-02".
func ExtractDateString(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Date == nil {
		return ""
	}
	return prop.Date.Start.Format("2006-01-02")
}

// ExtractDateFullString extracts a date property value using the DateTime.String() format.
// Used by Timesheet and TaskOrderLog for full date string representation.
func ExtractDateFullString(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Date == nil {
		return ""
	}
	return prop.Date.Start.String()
}

// ExtractFormulaString extracts a formula property's string result.
func ExtractFormulaString(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Formula == nil {
		return ""
	}
	if prop.Formula.String != nil {
		return *prop.Formula.String
	}
	return ""
}

// ExtractFormulaNumber extracts a formula property's number result.
func ExtractFormulaNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Formula == nil || prop.Formula.Number == nil {
		return 0
	}
	return *prop.Formula.Number
}

// ExtractRollupTitle extracts the title from the first item of a rollup array.
func ExtractRollupTitle(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return ""
	}
	if prop.Rollup.Type == "array" && len(prop.Rollup.Array) > 0 {
		firstItem := prop.Rollup.Array[0]
		if len(firstItem.Title) > 0 {
			var b strings.Builder
			for _, rt := range firstItem.Title {
				b.WriteString(rt.PlainText)
			}
			return b.String()
		}
	}
	return ""
}

// ExtractRollupText tries to get title, then rich_text from the first rollup array item.
// Used by RefundRequests for the "Person" rollup.
func ExtractRollupText(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return ""
	}
	if prop.Rollup.Type == "array" && len(prop.Rollup.Array) > 0 {
		firstItem := prop.Rollup.Array[0]
		if len(firstItem.Title) > 0 {
			var b strings.Builder
			for _, rt := range firstItem.Title {
				b.WriteString(rt.PlainText)
			}
			return b.String()
		}
		if len(firstItem.RichText) > 0 {
			var b strings.Builder
			for _, rt := range firstItem.RichText {
				b.WriteString(rt.PlainText)
			}
			return b.String()
		}
	}
	return ""
}

// ExtractRollupRichTextFirst extracts the first RichText plain text from the first rollup array item.
// Used by ContractorRates for the "Discord" rollup.
func ExtractRollupRichTextFirst(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return ""
	}
	for _, item := range prop.Rollup.Array {
		if len(item.RichText) > 0 {
			return item.RichText[0].PlainText
		}
	}
	return ""
}

// ExtractRollupRichTextAll extracts all RichText from all rollup array items, joined with newlines.
// Used by ContractorFees for "Proof of Works" rollup.
func ExtractRollupRichTextAll(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return ""
	}
	var b strings.Builder
	for _, item := range prop.Rollup.Array {
		for _, rt := range item.RichText {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(rt.PlainText)
		}
	}
	return b.String()
}

// ExtractRollupNumber extracts a number from a rollup property.
// Handles both direct number rollups and array rollups (summing all numbers).
func ExtractRollupNumber(props nt.DatabasePageProperties, propName string) float64 {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return 0
	}
	switch prop.Rollup.Type {
	case "number":
		if prop.Rollup.Number != nil {
			return *prop.Rollup.Number
		}
	case "array":
		var sum float64
		for _, item := range prop.Rollup.Array {
			if item.Number != nil {
				sum += *item.Number
			}
		}
		return sum
	}
	return 0
}

// ExtractRollupSelect extracts the first select value from a rollup array.
func ExtractRollupSelect(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil {
		return ""
	}
	for _, item := range prop.Rollup.Array {
		if item.Select != nil {
			return item.Select.Name
		}
	}
	return ""
}

// ExtractEmail extracts an email property value.
func ExtractEmail(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Email == nil {
		return ""
	}
	return *prop.Email
}

// ExtractEmailFromRollup extracts the first email from a rollup array.
func ExtractEmailFromRollup(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Rollup == nil || len(prop.Rollup.Array) == 0 {
		return ""
	}
	firstItem := prop.Rollup.Array[0]
	if firstItem.Email != nil {
		return *firstItem.Email
	}
	return ""
}
