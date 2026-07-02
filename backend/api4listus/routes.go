package api4listus

import (
	"net/http"

	"github.com/sneat-co/sneat-go-core/extension"
)

// RegisterHttpRoutes registers listus routes
func RegisterHttpRoutes(handle extension.HTTPHandleFunc) {
	handle(http.MethodPost, "/v0/listus/create_list", httpPostCreateList)
	handle(http.MethodDelete, "/v0/listus/delete_list", httpDeleteList)
	handle(http.MethodPost, "/v0/listus/list_items_create", httpPostCreateListItems)
	handle(http.MethodPost, "/v0/listus/list_items_set_is_done", httpPostSetListItemsIsDone)
	handle(http.MethodDelete, "/v0/listus/list_items_delete", httpDeleteListItems)
	handle(http.MethodPost, "/v0/listus/list_items_reorder", httpPostReorderListItem)
	handle(http.MethodPost, "/v0/listus/list_items_set_watch_with", httpPostSetListItemWatchWith)
	handle(http.MethodPost, "/v0/listus/movies/search", httpPostSearchMovies)
	handle(http.MethodPost, "/v0/listus/movies/resolve", httpPostResolveMovie)
	handle(http.MethodPost, "/v0/listus/movies/identify", httpPostIdentifyMovies)
	handle(http.MethodPost, "/v0/listus/movies/add_to_watchlist", httpPostAddMovieToWatchlist)
}
