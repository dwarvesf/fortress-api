package invoiceemail

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type InvoiceEmailHandler struct {
	DB              *Database
	InvoiceDetector *InvoiceDetector
}

type InvoiceEmail struct {
	ID            int       `json:"id"`
	Sender        string    `json:"sender"`
	Subject       string    `json:"subject"`
	ReceivedAt    time.Time `json:"received_at"`
	InvoiceNumber string    `json:"invoice_number"`
	InvoiceAmount float64   `json:"invoice_amount"`
	InvoiceDate   time.Time `json:"invoice_date"`
	Content       string    `json:"content"`
}

func (h *InvoiceEmailHandler) GetInvoiceEmails(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.DB.GetInvoiceEmails()
	if err != nil {
		http.Error(w, "Failed to get invoice emails", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(invoices)
}

func (h *InvoiceEmailHandler) GetInvoiceEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	invoice, err := h.DB.GetInvoiceEmail(id)
	if err != nil {
		http.Error(w, "Invoice email not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(invoice)
}

func (h *InvoiceEmailHandler) CreateInvoiceEmail(w http.ResponseWriter, r *http.Request) {
	var invoice InvoiceEmail
	err := json.NewDecoder(r.Body).Decode(&invoice)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if h.InvoiceDetector.DetectInvoice(invoice) {
		invoiceData := h.InvoiceDetector.ExtractInvoiceData(invoice)
		err = h.DB.SaveInvoice(invoiceData)
		if err != nil {
			http.Error(w, "Failed to save invoice", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(invoice)
	} else {
		http.Error(w, "Not a valid invoice email", http.StatusBadRequest)
	}
}

func (h *InvoiceEmailHandler) UpdateInvoiceEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var invoice InvoiceEmail
	err := json.NewDecoder(r.Body).Decode(&invoice)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.DB.UpdateInvoiceEmail(id, invoice)
	if err != nil {
		http.Error(w, "Failed to update invoice email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoice)
}

func (h *InvoiceEmailHandler) DeleteInvoiceEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.DB.DeleteInvoiceEmail(id)
	if err != nil {
		http.Error(w, "Failed to delete invoice email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}