package db

import (
	"context"
	"database/sql"
	"testing"
)

func TestIssuesCRUD(t *testing.T) {
	ctx := context.Background()
	db, stop := dut(ctx, t)
	defer stop()

	s := db.Do(ctx)

	// Nonexistent issue should fail
	_, err := s.Issue().Get(42)
	if want, got := IssueErrorNotFound, err; want != got {
		t.Fatalf("Issue.Get(nonexistent id): wanted %v, got %v", want, got)
	}

	// Issue creation happy path
	issue, err := s.Issue().New(&Issue{
		Reporter: "q3k",
		Title:    "test issue",
		Assignee: "implr",
		Type:     1,
		Priority: 3,
		Status:   2,
	})
	if err != nil {
		t.Fatalf("Issue.New(okay): wanted nil, got %v", err)
	}
	if issue.ID == 0 {
		t.Fatalf("Issue.New(okay) didn't set an issue id")
	}

	// Retrieve that issue now
	id := issue.ID
	issue, err = s.Issue().Get(id)
	if err != nil {
		t.Fatalf("Issue.Get(%d): wanted nil, got %v", id, err)
	}
	if want, got := "q3k", issue.Reporter; want != got {
		t.Errorf("issue.Reporter is %q, want %q", want, got)
	}
	if want, got := "test issue", issue.Title; want != got {
		t.Errorf("issue.Title is %q, want %q", want, got)
	}
	if want, got := "implr", issue.Assignee; want != got {
		t.Errorf("issue.Assignee is %q, want %q", want, got)
	}
	if want, got := int64(1), issue.Type; want != got {
		t.Errorf("issue.Type is %d, want %d", want, got)
	}
	if want, got := int64(3), issue.Priority; want != got {
		t.Errorf("issue.Priority is %d, want %d", want, got)
	}
	if want, got := int64(2), issue.Status; want != got {
		t.Errorf("issue.Status is %d, want %d", want, got)
	}
}

func TestIssueUpdates(t *testing.T) {
	ctx := context.Background()
	db, stop := dut(ctx, t)
	defer stop()

	s := db.Do(ctx)

	issue, err := s.Issue().New(&Issue{
		Reporter: "q3k",
		Title:    "test issue",
		Assignee: "implr",
		Type:     1,
		Priority: 3,
		Status:   2,
	})
	if err != nil {
		t.Fatalf("Issue.New(okay): wanted nil, got %v", err)
	}
	id := issue.ID

	err = s.Issue().Update(&IssueUpdate{
		IssueID: id,
		Author:  "q3k",
		Title:   sql.NullString{"better issue", true},
	})
	if err != nil {
		t.Fatalf("Issue.Update(title: 'better issue'): wanted nil, got %v", err)
	}

	// Check if issue got updates
	issue, err = s.Issue().Get(id)
	if err != nil {
		t.Fatalf("Issue.Get(%d): wanted nil, got %v", id, err)
	}
	if want, got := "better issue", issue.Title; want != got {
		t.Fatalf("Issue.Title: wanted %q, got %q", want, got)
	}
}

func TestIssueHistory(t *testing.T) {
	ctx := context.Background()
	db, stop := dut(ctx, t)
	defer stop()

	s := db.Do(ctx)

	issue, err := s.Issue().New(&Issue{
		Reporter: "q3k",
		Title:    "test issue",
		Assignee: "implr",
		Type:     1,
		Priority: 3,
		Status:   2,
	})
	if err != nil {
		t.Fatalf("Issue.New(okay): wanted nil, got %v", err)
	}

	for i, test := range []struct {
		title    string
		assignee string
	}{
		{"test issue - foo", ""},
		{"test issue - foo", "foo"},
		{"", "q3k"},
	} {
		u := &IssueUpdate{IssueID: issue.ID}
		if test.title != "" {
			u.Title = sql.NullString{test.title, true}
		}
		if test.assignee != "" {
			u.Assignee = sql.NullString{test.assignee, true}
		}
		err := s.Issue().Update(u)
		if err != nil {
			t.Fatalf("test %d: %v", i, err)
		}

		updates, err := s.Issue().GetHistory(issue.ID, &IssueGetHistoryOpts{})
		if err != nil {
			t.Fatalf("test %d: GetHistory: %v", i, err)
		}

		if want, get := i+1, len(updates); want < get {
			t.Fatalf("test %d: wanted %d updates, got %d", want, get)
		}

		update := updates[i]
		if test.title != "" {
			if want, get := test.title, update.Title.String; want != get {
				t.Errorf("test %d, title, wanted %q, got %q", i, want, get)
			}
		}
		if test.assignee != "" {
			if want, get := test.assignee, update.Assignee.String; want != get {
				t.Errorf("test %d, assignee, wanted %q, got %q", i, want, get)
			}
		}

		issue2, err := s.Issue().Get(issue.ID)
		if err != nil {
			t.Fatalf("test %d, Issue().Get(%d): %v", i, issue.ID, err)
		}

		if test.title != "" {
			if want, get := test.title, issue2.Title; want != get {
				t.Errorf("test %d, title, wanted %q got %q", i, want, get)
			}
		}
		if test.assignee != "" {
			if want, get := test.assignee, issue2.Assignee; want != get {
				t.Errorf("test %d, assignee, wanted %q got %q", i, want, get)
			}
		}
	}
}
