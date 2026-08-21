package db

import (
	"database/sql"
	"errors"
	"testing"
)

func TestCreateJobApplication(t *testing.T) {
	notes := "Referred by a friend"
	app, err := CreateJobApplication(testDB, "Acme Corp", "Backend Engineer", &notes)
	if err != nil {
		t.Fatalf("CreateJobApplication returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	if app.ID == 0 {
		t.Error("expected a generated ID, got 0")
	}
	if app.Company != "Acme Corp" {
		t.Errorf("Company = %q, want %q", app.Company, "Acme Corp")
	}
	if app.RoleTitle != "Backend Engineer" {
		t.Errorf("RoleTitle = %q, want %q", app.RoleTitle, "Backend Engineer")
	}
	if app.Notes == nil || *app.Notes != notes {
		t.Errorf("Notes = %v, want %q", app.Notes, notes)
	}

	var status string
	err = testDB.QueryRow(
		`SELECT status FROM application_events WHERE job_application_id = $1`,
		app.ID,
	).Scan(&status)
	if err != nil {
		t.Fatalf("querying initial event: %v", err)
	}
	if status != "applied" {
		t.Errorf("initial event status = %q, want %q", status, "applied")
	}
}

func TestListJobApplications(t *testing.T) {

	app, err := CreateJobApplication(testDB, "ByteDance", "SRE Intern", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	apps, err := ListJobApplications(testDB)
	if err != nil {
		t.Fatalf("ListJobApplications returned error: %v", err)
	}

	var found *JobApplicationSummary
	for i := range apps {
		if apps[i].ID == app.ID {
			found = &apps[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("created application (id=%d) not found in list", app.ID)
	}

	if found.Company != "ByteDance" {
		t.Errorf("Company = %q, want %q", found.Company, "Globex")
	}
	if found.CurrentStatus == nil || *found.CurrentStatus != "applied" {
		t.Errorf("CurrentStatus = %v, want %q", found.CurrentStatus, "applied")
	}
	if found.CurrentRound != nil {
		t.Errorf("CurrentRound = %v, want nil", found.CurrentRound)
	}

}

func TestGetJobApplication(t *testing.T) {
	app, err := CreateJobApplication(testDB, "Initech", "Platform Engineer", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	detail, err := GetJobApplication(testDB, app.ID)
	if err != nil {
		t.Fatalf("GetJobApplication returned error: %v", err)
	}

	if detail.Company != "Initech" {
		t.Errorf("Company = %q, want %q", detail.Company, "Initech")
	}
	if detail.CurrentStatus == nil || *detail.CurrentStatus != "applied" {
		t.Errorf("CurrentStatus = %v, want %q", detail.CurrentStatus, "applied")
	}
	if len(detail.Events) != 1 {
		t.Fatalf("len(Events) = %d, want 1", len(detail.Events))
	}
	if detail.Events[0].Status != "applied" {
		t.Errorf("Events[0].Status = %q, want %q", detail.Events[0].Status, "applied")
	}
}

func TestGetJobApplication_NotFound(t *testing.T) {
	_, err := GetJobApplication(testDB, -1)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestUpdateJobApplication(t *testing.T) {
	app, err := CreateJobApplication(testDB, "VISA", "Software Engineer Intern", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication returned error: %v", err)
	}

	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	newNotes := "Updated after phone screen"
	updated, err := UpdateJobApplication(testDB, app.ID, nil, nil, &newNotes)
	if updated.Company != "VISA" {
		t.Errorf("Company = %q, want %q (should be unchanged)", updated.Company, "VISA")
	}
	if updated.RoleTitle != "Software Engineer Intern" {
		t.Errorf("RoleTitle = %q, want %q (should be unchanged)", updated.RoleTitle, "Software Engineer Intern")
	}
	if updated.Notes == nil || *updated.Notes != newNotes {
		t.Errorf("Notes = %v, want %q", updated.Notes, newNotes)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Errorf("UpdatedAt (%v) should be after CreatedAt (%v)", updated.UpdatedAt, updated.CreatedAt)
	}
}

func TestUpdateJobApplication_NotFound(t *testing.T) {
	notes := "doesn't matter"
	_, err := UpdateJobApplication(testDB, -1, nil, nil, &notes)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestDeleteJobApplication(t *testing.T) {
	app, err := CreateJobApplication(testDB, "Wayne Corp", "DevOps Engineer", nil)
	if err != nil {
		t.Fatalf("CreateJobApplication returned error: %v", err)
	}
	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM job_applications WHERE id = $1`, app.ID)
	})

	if _, err := CreateApplicationEvent(testDB, app.ID, "interviewing", nil); err != nil {
		t.Fatalf("CreateApplicationEvent returned error: %v", err)
	}

	if err := DeleteJobApplication(testDB, app.ID); err != nil {
		t.Fatalf("DeleteJobApplication returned error: %v", err)
	}

	_, err = GetJobApplication(testDB, app.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("GetJobApplication after delete: err = %v, want sql.ErrNoRows", err)
	}

	var count int
	err = testDB.QueryRow(`SELECT count(*) FROM application_events WHERE job_application_id = $1`, app.ID).Scan(&count)
	if err != nil {
		t.Fatalf("querying application_events count: %v", err)
	}
	if count != 0 {
		t.Errorf("application_events count after delete = %d, want 0 (cascade should remove them)", count)
	}
}

func TestDeleteJobApplication_NotFound(t *testing.T) {
	err := DeleteJobApplication(testDB, -1)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("err = %v, want sql.ErrNoRows", err)
	}
}
