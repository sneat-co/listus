package api4listus

import (
	"net/http"

	calendarfacade "github.com/sneat-co/sneat-ext-contracts/calendarius/facade4calendarius"
	"github.com/sneat-co/sneat-go-core/extension"
)

type Dependencies struct {
	SourceLinkedDateTasks calendarfacade.SourceLinkedDateTaskProvider
}

// RegisterHttpRoutes registers listus routes
func RegisterHttpRoutes(handle extension.HTTPHandleFunc, dependencies ...Dependencies) {
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
	if len(dependencies) > 0 && dependencies[0].SourceLinkedDateTasks != nil {
		handle(http.MethodPost, "/v0/listus/item_date_task_save", httpPostSaveListItemDateTask(dependencies[0].SourceLinkedDateTasks))
	}
}
