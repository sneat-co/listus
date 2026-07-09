package facade4listus

import (
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/listus/backend/dbo4listus"
)

func TestGetList_ReturnsListWithItems(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)

	// Seed a standard list with one item via CreateListItems.
	createItems(t, dbo4listus.DoTasksListID, "Task A")

	list, err := GetList(userCtx(testUserID), listRequest(testSpaceID, dbo4listus.DoTasksListID))
	if err != nil {
		t.Fatalf("GetList failed: %v", err)
	}
	if !list.Record.Exists() {
		t.Fatal("expected list record to exist")
	}
	if len(list.Data.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(list.Data.Items))
	}
	if list.Data.Items[0].Title != "Task A" {
		t.Errorf("expected item title %q, got %q", "Task A", list.Data.Items[0].Title)
	}
}

func TestGetList_NonExistentStandardList_ReturnsNotFound(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)

	// DoTasksListID is a standard list that has never been written — should be not found.
	_, err := GetList(userCtx(testUserID), listRequest(testSpaceID, dbo4listus.DoTasksListID))
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
	if !dal.IsNotFound(err) {
		t.Errorf("expected IsNotFound(err) == true, got err=%v", err)
	}
}

func TestGetList_NonMemberUser_ReturnsError(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)

	// Seed the list as the member user.
	createItems(t, dbo4listus.DoTasksListID, "Task A")

	// A user that is not in the space's UserIDs should be rejected.
	_, err := GetList(userCtx("non-member-user"), listRequest(testSpaceID, dbo4listus.DoTasksListID))
	if err == nil {
		t.Fatal("expected error for non-member user, got nil")
	}
}

func TestGetList_InvalidRequest_ReturnsValidationError(t *testing.T) {
	_ = newTestDBWithSpace(t, testSpaceID, testUserID)

	// Empty ListID should fail validation before hitting the DB.
	_, err := GetList(userCtx(testUserID), listRequest(testSpaceID, ""))
	if err == nil {
		t.Error("expected validation error for empty listID")
	}
}
