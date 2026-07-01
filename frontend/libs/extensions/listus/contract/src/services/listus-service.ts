import { InjectionToken } from '@angular/core';
import { ISpaceContext } from '@sneat/space-models';
import { Observable } from 'rxjs';
import { IListContext } from '../contexts';
import { ListType } from '../dto';
import {
  AddMovieToWatchlistRequest,
  AddMovieToWatchlistResponse,
  ICreateListRequest,
  IDeleteListItemsRequest,
  IListItemResult,
  IListItemsCommandParams,
  IReorderListItemsRequest,
  ISetListItemsIsComplete,
  ResolveMovieRequest,
  ResolveMovieResponse,
  SearchMoviesRequest,
  SearchMoviesResponse,
  SetListItemWatchWithRequest,
} from './interfaces';

// IListusService is the runtime-light contract the listus pages and components
// depend on. Members mirror the concrete ListService public surface used by the
// UI; the implementation lives in the internal lib and is provided via the
// LISTUS_SERVICE token below. The shared BaseListPage additionally needs the
// inherited ModuleSpaceItemService surface, so it types the injected token as
// an intersection with ModuleSpaceItemService<IListBrief, IListDbo>.
export interface IListusService {
  createList(request: ICreateListRequest): Observable<IListContext>;
  deleteList(space: ISpaceContext, listId: string): Observable<void>;
  reorderListItems(request: IReorderListItemsRequest): Observable<void>;
  createListItems(
    params: IListItemsCommandParams,
  ): Observable<IListItemResult>;
  setListItemsIsCompleted(
    request: ISetListItemsIsComplete,
  ): Observable<void>;
  deleteListItems(request: IDeleteListItemsRequest): Observable<void>;
  getListById(
    space: ISpaceContext,
    listType: ListType,
    listID: string,
  ): Observable<IListContext>;
  // Movie search/resolve are read-only TMDB proxies (see listus/movies/search
  // & listus/movies/resolve); addMovieToWatchlist resolves a movie server-side
  // and appends it (fully enriched) to the space's canonical watch!movies list.
  searchMovies(request: SearchMoviesRequest): Observable<SearchMoviesResponse>;
  resolveMovie(request: ResolveMovieRequest): Observable<ResolveMovieResponse>;
  addMovieToWatchlist(
    request: AddMovieToWatchlistRequest,
  ): Observable<AddMovieToWatchlistResponse>;
  setListItemWatchWith(request: SetListItemWatchWithRequest): Observable<void>;
}

export const LISTUS_SERVICE = new InjectionToken<IListusService>(
  'ListusService',
);
