package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jobtracedb "github.com/ZadeNova/jobtrace/internal/db"
)

func TestCreateJobApplicationHandler(t *testing.T) {
	handler := CreateJobApplicationHandler(testDB)

	req := httptest.NewRequest("POST", "/job-applications", strings.NewReader(`{"company":"Acme","role_title":"SRE"}`))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var app jobtracedb.JobApplication
	if err := json.Unmarshal(rec.Body.Bytes(), &app); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	if app.Company != "Acme" {
		t.Errorf("Company = %q, want %q", app.Company, "Acme")
	}
}

func TestCreateJobApplicationHandler_MissingCompany(t *testing.T) {
	handler := CreateJobApplicationHandler(testDB)

	req := httptest.NewRequest("POST", "/job-applications", strings.NewReader(`{"role_title":"SRE"}`))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
