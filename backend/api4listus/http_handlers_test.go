package api4listus

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sneat-co/listus/backend/dal4listus"
	"github.com/sneat-co/listus/backend/dbo4listus"
	"github.com/sneat-co/listus/backend/dto4listus"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/facade"
)

// --- Auth bypass scaffolding (mirrors assetus api handler tests) -------------

type mockUserContext struct {
	facade.UserContext
	userID string
}

func (m mockUserContext) GetUserID() string { return m.userID }

type mockContextWithUser struct {
	facade.ContextWithUser
	ctx  context.Context
	user facade.UserContext
}

func (m mockContextWithUser) User() facade.UserContext { return m.user }
func (m mockContextWithUser) Value(key any) any        { return m.ctx.Value(key) }

func authAsUser(t *testing.T) {
	t.Helper()
	old := apicore.VerifyRequestAndCreateUserContext
	apicore.VerifyRequestAndCreateUserContext = func(w http.ResponseWriter, r *http.Request, options verify.RequestOptions) (facade.ContextWithUser, error) {
		return mockContextWithUser{ctx: t.Context(), user: mockUserContext{userID: "u1"}}, nil
	}
	t.Cleanup(func() { apicore.VerifyRequestAndCreateUserContext = old })
}

func authRejected(t *testing.T) {
	t.Helper()
	old := apicore.VerifyRequestAndCreateUserContext
	apicore.VerifyRequestAndCreateUserContext = func(w http.ResponseWriter, r *http.Request, options verify.RequestOptions) (facade.ContextWithUser, error) {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, facade.ErrUnauthenticated
	}
	t.Cleanup(func() { apicore.VerifyRequestAndCreateUserContext = old })
}

func newPostRequest(path, body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
}

const listQuery = "spaceID=s1&listID=do!tasks"

var errBoom = errors.New("boom")

// =============================================================================
// Routes
// =============================================================================

func TestRegisterHttpRoutes(t *testing.T) {
	type reg struct{ method, path string }
	var got []reg
	handle := func(method, path string, _ http.HandlerFunc) {
		if !strings.HasPrefix(path, "/v0/listus/") {
			t.Errorf("path %q does not start with /v0/listus/", path)
		}
		got = append(got, reg{method, path})
	}
	RegisterHttpRoutes(handle)
	want := []reg{
		{http.MethodPost, "/v0/listus/create_list"},
		{http.MethodDelete, "/v0/listus/delete_list"},
		{http.MethodPost, "/v0/listus/list_items_create"},
		{http.MethodPost, "/v0/listus/list_items_set_is_done"},
		{http.MethodDelete, "/v0/listus/list_items_delete"},
		{http.MethodPost, "/v0/listus/list_items_reorder"},
	}
	if len(got) != len(want) {
		t.Fatalf("registered %d routes, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("route %d = %+v, want %+v", i, got[i], w)
		}
	}
}

// =============================================================================
// httpPostCreateList
// =============================================================================

func TestHttpPostCreateList_201OnSuccess(t *testing.T) {
	authAsUser(t)
	old := createList
	t.Cleanup(func() { createList = old })
	createList = func(ctx facade.ContextWithUser, request dto4listus.CreateListRequest) (dto4listus.CreateListResponse, error) {
		return dto4listus.CreateListResponse{ID: "new-list"}, nil
	}
	w := httptest.NewRecorder()
	httpPostCreateList(w, newPostRequest("/v0/listus/create_list", `{"spaceID":"s1","type":"do","title":"My List"}`))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "new-list") {
		t.Errorf("body %q missing created id", w.Body.String())
	}
}

func TestHttpPostCreateList_500WhenFacadeFails(t *testing.T) {
	authAsUser(t)
	old := createList
	t.Cleanup(func() { createList = old })
	createList = func(ctx facade.ContextWithUser, request dto4listus.CreateListRequest) (dto4listus.CreateListResponse, error) {
		return dto4listus.CreateListResponse{}, errBoom
	}
	w := httptest.NewRecorder()
	httpPostCreateList(w, newPostRequest("/v0/listus/create_list", `{"spaceID":"s1","type":"do","title":"My List"}`))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", w.Code, w.Body.String())
	}
}

func TestHttpPostCreateList_400OnBadJSON(t *testing.T) {
	authAsUser(t)
	w := httptest.NewRecorder()
	httpPostCreateList(w, newPostRequest("/v0/listus/create_list", "not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

func TestHttpPostCreateList_401WhenUnauthenticated(t *testing.T) {
	authRejected(t)
	w := httptest.NewRecorder()
	httpPostCreateList(w, newPostRequest("/v0/listus/create_list", `{"spaceID":"s1","type":"do","title":"X"}`))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

// =============================================================================
// httpDeleteList
// =============================================================================

func TestHttpDeleteList_201OnSuccess(t *testing.T) {
	authAsUser(t)
	old := deleteList
	t.Cleanup(func() { deleteList = old })
	var gotReq dto4listus.ListRequest
	deleteList = func(ctx facade.ContextWithUser, request dto4listus.ListRequest) error {
		gotReq = request
		return nil
	}
	req := httptest.NewRequest(http.MethodDelete, "/v0/listus/delete_list?"+listQuery, nil)
	w := httptest.NewRecorder()
	httpDeleteList(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
	if string(gotReq.SpaceID) != "s1" || string(gotReq.ListID) != "do!tasks" {
		t.Errorf("facade got %+v, want spaceID=s1 listID=do!tasks", gotReq)
	}
}

func TestHttpDeleteList_500WhenFacadeFails(t *testing.T) {
	authAsUser(t)
	old := deleteList
	t.Cleanup(func() { deleteList = old })
	deleteList = func(ctx facade.ContextWithUser, request dto4listus.ListRequest) error { return errBoom }
	req := httptest.NewRequest(http.MethodDelete, "/v0/listus/delete_list?"+listQuery, nil)
	w := httptest.NewRecorder()
	httpDeleteList(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", w.Code, w.Body.String())
	}
}

func TestHttpDeleteList_401WhenUnauthenticated(t *testing.T) {
	authRejected(t)
	req := httptest.NewRequest(http.MethodDelete, "/v0/listus/delete_list?"+listQuery, nil)
	w := httptest.NewRecorder()
	httpDeleteList(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

// =============================================================================
// httpPostCreateListItems
// =============================================================================

func TestHttpPostCreateListItems_201OnSuccess(t *testing.T) {
	authAsUser(t)
	old := createListItems
	t.Cleanup(func() { createListItems = old })
	createListItems = func(ctx facade.ContextWithUser, request dto4listus.CreateListItemsRequest) (dto4listus.CreateListItemResponse, dal4listus.ListEntry, error) {
		return dto4listus.CreateListItemResponse{CreatedItems: []*dbo4listus.ListItemBrief{{ID: "it-1"}}}, dal4listus.ListEntry{}, nil
	}
	body := `{"spaceID":"s1","listID":"do!tasks","items":[{"title":"Milk"}]}`
	w := httptest.NewRecorder()
	httpPostCreateListItems(w, newPostRequest("/v0/listus/list_items_create", body))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "it-1") {
		t.Errorf("body %q missing item id", w.Body.String())
	}
}

func TestHttpPostCreateListItems_500WhenFacadeFails(t *testing.T) {
	authAsUser(t)
	old := createListItems
	t.Cleanup(func() { createListItems = old })
	createListItems = func(ctx facade.ContextWithUser, request dto4listus.CreateListItemsRequest) (dto4listus.CreateListItemResponse, dal4listus.ListEntry, error) {
		return dto4listus.CreateListItemResponse{}, dal4listus.ListEntry{}, errBoom
	}
	body := `{"spaceID":"s1","listID":"do!tasks","items":[{"title":"Milk"}]}`
	w := httptest.NewRecorder()
	httpPostCreateListItems(w, newPostRequest("/v0/listus/list_items_create", body))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", w.Code, w.Body.String())
	}
}

func TestHttpPostCreateListItems_400OnBadJSON(t *testing.T) {
	authAsUser(t)
	w := httptest.NewRecorder()
	httpPostCreateListItems(w, newPostRequest("/v0/listus/list_items_create", "not json"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

// =============================================================================
// httpDeleteListItems
// =============================================================================

func TestHttpDeleteListItems_201OnSuccess(t *testing.T) {
	authAsUser(t)
	old := deleteListItems
	t.Cleanup(func() { deleteListItems = old })
	deleteListItems = func(ctx facade.ContextWithUser, request dto4listus.ListItemIDsRequest) ([]*dbo4listus.ListItemBrief, dal4listus.ListEntry, error) {
		return nil, dal4listus.ListEntry{}, nil
	}
	body := `{"itemIDs":["a"]}`
	w := httptest.NewRecorder()
	httpDeleteListItems(w, newPostRequest("/v0/listus/list_items_delete?"+listQuery, body))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
}

func TestHttpDeleteListItems_500WhenFacadeFails(t *testing.T) {
	authAsUser(t)
	old := deleteListItems
	t.Cleanup(func() { deleteListItems = old })
	deleteListItems = func(ctx facade.ContextWithUser, request dto4listus.ListItemIDsRequest) ([]*dbo4listus.ListItemBrief, dal4listus.ListEntry, error) {
		return nil, dal4listus.ListEntry{}, errBoom
	}
	body := `{"itemIDs":["a"]}`
	w := httptest.NewRecorder()
	httpDeleteListItems(w, newPostRequest("/v0/listus/list_items_delete?"+listQuery, body))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", w.Code, w.Body.String())
	}
}

// =============================================================================
// httpPostSetListItemsIsDone
// =============================================================================

func TestHttpPostSetListItemsIsDone_204OnSuccess(t *testing.T) {
	authAsUser(t)
	old := setListItemsIsDone
	t.Cleanup(func() { setListItemsIsDone = old })
	setListItemsIsDone = func(ctx facade.ContextWithUser, request dto4listus.ListItemsSetIsDoneRequest) ([]*dbo4listus.ListItemBrief, dal4listus.ListEntry, error) {
		return nil, dal4listus.ListEntry{}, nil
	}
	body := `{"itemIDs":["a"],"isDone":true}`
	w := httptest.NewRecorder()
	httpPostSetListItemsIsDone(w, newPostRequest("/v0/listus/list_items_set_is_done?"+listQuery, body))
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body=%s", w.Code, w.Body.String())
	}
}

func TestHttpPostSetListItemsIsDone_500WhenFacadeFails(t *testing.T) {
	authAsUser(t)
	old := setListItemsIsDone
	t.Cleanup(func() { setListItemsIsDone = old })
	setListItemsIsDone = func(ctx facade.ContextWithUser, request dto4listus.ListItemsSetIsDoneRequest) ([]*dbo4listus.ListItemBrief, dal4listus.ListEntry, error) {
		return nil, dal4listus.ListEntry{}, errBoom
	}
	body := `{"itemIDs":["a"],"isDone":true}`
	w := httptest.NewRecorder()
	httpPostSetListItemsIsDone(w, newPostRequest("/v0/listus/list_items_set_is_done?"+listQuery, body))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", w.Code, w.Body.String())
	}
}

// =============================================================================
// httpPostReorderListItem
// =============================================================================

// httpPostReorderListItem passes a nil response with http.StatusOK to
// apicore.ReturnJSON, which panics (nil response is only valid with 204). This
// documents that the success path currently panics.
func TestHttpPostReorderListItem_SuccessPanicsViaReturnJSON(t *testing.T) {
	authAsUser(t)
	old := reorderListItem
	t.Cleanup(func() { reorderListItem = old })
	reorderListItem = func(ctx facade.ContextWithUser, request dto4listus.ReorderListItemsRequest) error { return nil }
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected ReturnJSON to panic for nil response with 200 status")
		}
	}()
	body := `{"itemIDs":["a"],"toIndex":0}`
	httpPostReorderListItem(httptest.NewRecorder(), newPostRequest("/v0/listus/list_items_reorder?"+listQuery, body))
}

func TestHttpPostReorderListItem_500WhenFacadeFails(t *testing.T) {
	authAsUser(t)
	old := reorderListItem
	t.Cleanup(func() { reorderListItem = old })
	reorderListItem = func(ctx facade.ContextWithUser, request dto4listus.ReorderListItemsRequest) error { return errBoom }
	body := `{"itemIDs":["a"],"toIndex":0}`
	w := httptest.NewRecorder()
	httpPostReorderListItem(w, newPostRequest("/v0/listus/list_items_reorder?"+listQuery, body))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", w.Code, w.Body.String())
	}
}
