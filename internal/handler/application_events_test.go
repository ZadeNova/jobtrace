package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	jobtracedb "github.com/ZadeNova/jobtrace/internal/db"
)

func TestCreateApplicationEventHandler_InvalidStatus(t *testing.T) {
	app, err := jobtracedb.CreateJobApplication(testDB, "Acme", "SRE", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	handler := CreateApplicationEventHandler(testDB)
	idStr := strconv.FormatInt(app.ID, 10)

	req := httptest.NewRequest("POST", "/job-applications/"+idStr+"/events", strings.NewReader(`{"status":"banana"}`))
	req.SetPathValue("id", idStr)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateApplicationEventHandler_NotFound(t *testing.T) {
	handler := CreateApplicationEventHandler(testDB)

	req := httptest.NewRequest("POST", "/job-applications/-1/events", strings.NewReader(`{"status":"applied"}`))
	req.SetPathValue("id", "-1")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

/*
Prevents a duplicate round number for a single job application. In case a race condition occurs, where eg 2 requests are sent at almost the same time.
The unique index in postgres prevents this too.
This test case is done to check to see if the hTTP request returned is correct.
*/
func TestCreateApplicationEventHandler_RoundConflict(t *testing.T) {
	app, err := jobtracedb.CreateJobApplication(testDB, "Conflict Co", "Test Role", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	handler := CreateApplicationEventHandler(testDB)
	idStr := strconv.FormatInt(app.ID, 10)

	const n = 10
	codes := make(chan int, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("POST", fmt.Sprintf("/job-applications/%s/events", idStr), strings.NewReader(`{"status":"interviewing"}`))
			req.SetPathValue("id", idStr)
			rec := httptest.NewRecorder()
			handler(rec, req)
			codes <- rec.Code
		}()
	}
	wg.Wait()
	close(codes)

	var created, conflicts int
	for code := range codes {
		switch code {
		case http.StatusCreated:
			created++
		case http.StatusConflict:
			conflicts++
		default:
			t.Errorf("unexpected status code: %d", code)
		}
	}

	if created == 0 {
		t.Error("expected at least one successful creation")
	}
	if conflicts == 0 {
		t.Error("expected at least one 409 from the concurrent race")
	}
}
