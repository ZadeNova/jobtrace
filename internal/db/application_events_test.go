package db

import (
	"database/sql"
	"errors"
	"testing"
)

func TestCreateApplicationEvent(t *testing.T) {
	app, err := CreateJobApplication(testDB, "Avengers Corp", "Software Engineer Intern", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	event, err := CreateApplicationEvent(testDB, app.ID, "rejected", nil)
	if err != nil {
		t.Fatalf("CreateApplicationEvent returned error: %v", err)
	}

	if event.Status != "rejected" {
		t.Errorf("Status = %q, want %q", event.Status, "rejected")
	}
	if event.RoundNumber != nil {
		t.Errorf("RoundNumber = %v, want nil", event.RoundNumber)
	}
	if event.JobApplicationID != app.ID {
		t.Errorf("JobApplicationID = %d, want %d", event.JobApplicationID, app.ID)
	}
}

func TestCreateApplicationEvent_RoundIncrement(t *testing.T) {
	app, err := CreateJobApplication(testDB, "LexCorp", "Site Reliability Engineer", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	first, err := CreateApplicationEvent(testDB, app.ID, "interviewing", nil)
	if err != nil {
		t.Fatalf("first CreateApplicationEvent returned error: %v", err)
	}
	if first.RoundNumber == nil || *first.RoundNumber != 1 {
		t.Errorf("first RoundNumber = %v, want 1", first.RoundNumber)
	}

	second, err := CreateApplicationEvent(testDB, app.ID, "interviewing", nil)
	if err != nil {
		t.Fatalf("second CreateApplicationEvent returned error: %v", err)
	}
	if second.RoundNumber == nil || *second.RoundNumber != 2 {
		t.Errorf("second RoundNumber = %v, want 2", second.RoundNumber)
	}
}

func TestCreateApplicationEvent_NotFound(t *testing.T) {
	_, err := CreateApplicationEvent(testDB, -1, "applied", nil)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestDeleteApplicationEvent(t *testing.T) {
	app, err := CreateJobApplication(testDB, "Oscorp", "Research Scientist", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	event, err := CreateApplicationEvent(testDB, app.ID, "interviewing", nil)
	if err != nil {
		t.Fatalf("CreateApplicationEvent returned error: %v", err)
	}

	if err := DeleteApplicationEvent(testDB, app.ID, event.ID); err != nil {
		t.Fatalf("DeleteApplicationEvent returned error: %v", err)
	}

	var count int
	err = testDB.QueryRow(`SELECT count(*) FROM application_events WHERE id = $1`, event.ID).Scan(&count)
	if err != nil {
		t.Fatalf("querying event count: %v", err)
	}
	if count != 0 {
		t.Errorf("event count after delete = %d, want 0", count)
	}
}

func TestDeleteApplicationEvent_NotFound(t *testing.T) {
	err := DeleteApplicationEvent(testDB, -1, -1)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestDeleteApplicationEvent_WrongApplication(t *testing.T) {
	appA, err := CreateJobApplication(testDB, "Company A", "Role A", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication (A) returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, appA.ID)
	})

	appB, err := CreateJobApplication(testDB, "Company B", "Role B", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication (B) returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, appB.ID)
	})

	eventA, err := CreateApplicationEvent(testDB, appA.ID, "interviewing", nil)
	if err != nil {
		t.Fatalf("CreateApplicationEvent returned error: %v", err)
	}

	err = DeleteApplicationEvent(testDB, appB.ID, eventA.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("err = %v, want sql.ErrNoRows", err)
	}

	var count int
	err = testDB.QueryRow(`SELECT count(*) FROM application_events WHERE id = $1`, eventA.ID).Scan(&count)
	if err != nil {
		t.Fatalf("querying event count: %v", err)
	}
	if count != 1 {
		t.Errorf("event count = %d, want 1 (should not have been deleted)", count)
	}
}
