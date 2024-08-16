package invoiceemail

import (
	"database/sql"
	"log"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(connectionString string) (*Database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}

func (d *Database) SaveInvoice(invoiceData map[string]interface{}) error {
	_, err := d.db.Exec(`
		INSERT INTO invoice_emails (sender, subject, received_at, invoice_number, invoice_amount, invoice_date, content)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, invoiceData["sender"], invoiceData["subject"], invoiceData["received_at"], invoiceData["invoice_number"], invoiceData["invoice_amount"], invoiceData["invoice_date"], invoiceData["content"])

	return err
}

func (d *Database) GetInvoiceEmails() ([]InvoiceEmail, error) {
	rows, err := d.db.Query("SELECT * FROM invoice_emails")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []InvoiceEmail
	for rows.Next() {
		var i InvoiceEmail
		err := rows.Scan(&i.ID, &i.Sender, &i.Subject, &i.ReceivedAt, &i.InvoiceNumber, &i.InvoiceAmount, &i.InvoiceDate, &i.Content)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, i)
	}

	return invoices, nil
}

func (d *Database) GetInvoiceEmail(id string) (InvoiceEmail, error) {
	var i InvoiceEmail
	err := d.db.QueryRow("SELECT * FROM invoice_emails WHERE id = $1", id).Scan(&i.ID, &i.Sender, &i.Subject, &i.ReceivedAt, &i.InvoiceNumber, &i.InvoiceAmount, &i.InvoiceDate, &i.Content)
	return i, err
}

func (d *Database) UpdateInvoiceEmail(id string, invoice InvoiceEmail) error {
	_, err := d.db.Exec(`
		UPDATE invoice_emails
		SET sender = $1, subject = $2, received_at = $3, invoice_number = $4, invoice_amount = $5, invoice_date = $6, content = $7
		WHERE id = $8
	`, invoice.Sender, invoice.Subject, invoice.ReceivedAt, invoice.InvoiceNumber, invoice.InvoiceAmount, invoice.InvoiceDate, invoice.Content, id)

	return err
}

func (d *Database) DeleteInvoiceEmail(id string) error {
	_, err := d.db.Exec("DELETE FROM invoice_emails WHERE id = $1", id)
	return err
}

func (d *Database) Close() error {
	return d.db.Close()
}