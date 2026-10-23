import { HttpParams } from '@angular/common/http';
import { Injectable, inject, Injector } from '@angular/core';
import {
  Firestore as AngularFirestore,
  CollectionReference,
  DocumentReference,
  collection,
  doc,
} from 'firebase/firestore';
import { SneatApiService, SneatFirestoreService } from '@sneat/api';
import { ISpaceContext } from '@sneat/space-models';
import { ModuleSpaceItemService } from '@sneat/space-services';
import { Observable, of, throwError } from 'rxjs';
import { map } from 'rxjs/operators';
import {
  AddMovieToWatchlistRequest,
  AddMovieToWatchlistResponse,
  IListBrief,
  IListContext,
  IListDbo,
  IListGroup,
  IListusSpaceDbo,
  ListType,
  ICreateListItemsRequest,
  ICreateListRequest,
  IDeleteListItemsRequest,
  IdentifyMoviesRequest,
  IdentifyMoviesResponse,
  IListItemResult,
  IListItemsCommandParams,
  IReorderListItemsRequest,
  ISetListItemsIsComplete,
  ISaveListItemDateTaskRequest,
  ISaveListItemDateTaskResponse,
  ResolveMovieRequest,
  ResolveMovieResponse,
  SearchMoviesRequest,
  SearchMoviesResponse,
  SetListItemWatchWithRequest,
  IApplyListTemplateRequest,
  IApplyListTemplateResult,
} from '@sneat/extension-listus-contract';

@Injectable()
export class ListService extends ModuleSpaceItemService<IListBrief, IListDbo> {
  private readonly listusSpaceRecords: SneatFirestoreService<
    IListusSpaceDbo,
    IListusSpaceDbo
  >;

  constructor() {
    const afs = inject(AngularFirestore);
    const sneatApiService = inject(SneatApiService);
    const injector = inject(Injector);
    super(injector, 'listus', 'lists', afs, sneatApiService);
    this.listusSpaceRecords = new SneatFirestoreService(injector);
  }

  public observeSpaceLists(
    spaceID: string,
  ): Observable<Readonly<Record<string, IListBrief>>> {
    const extensionRecords = collection(
      this.spacesCollection,
      spaceID,
      'ext',
    ) as CollectionReference<IListusSpaceDbo>;
    return this.listusSpaceRecords
      .watchByID(extensionRecords, 'listus')
      .pipe(map((record) => record.dbo?.lists || {}));
  }

  public createList(request: ICreateListRequest): Observable<IListContext> {
    return this.sneatApiService
      .post<IListContext>('listus/create_list', request)
      .pipe(
        map((list) => {
          if (!list.brief) {
            if (!list.dbo) {
              throw new Error('response does not have either brief or dto');
            }
            const title = list.dbo.title || list.id;
            list = { ...list, brief: { ...list.dbo, title } };
          }
          return list;
        }),
      );
  }

  public deleteList(space: ISpaceContext, listId: string): Observable<void> {
    if (space.type === 'family' && listId === 'groceries') {
      return throwError(
        () => 'groceries list is not removable for family team',
      );
    }
    return this.sneatApiService.delete(
      `lists/delete_list?space=${space.id}&id=${listId}`,
    );
  }

  public reorderListItems(request: IReorderListItemsRequest): Observable<void> {
    return this.sneatApiService.post(`listus/list_items_reorder?`, request);
  }

  private getFullListID(type: ListType, shortID: string): string {
    return `${type}!${shortID}`;
  }

  public createListItems(
    params: IListItemsCommandParams,
  ): Observable<IListItemResult> {
    if (!params.list.type) {
      return throwError(() => 'list is of unknown type: ' + params.list.type);
    }
    const request: ICreateListItemsRequest = {
      spaceID: params.space.id,
      listID: params.list.id,
      // listType: params.list.type,
      items: params.items,
    };
    return this.sneatApiService.post('listus/list_items_create', request);
  }

  public applyListTemplate(
    request: IApplyListTemplateRequest,
  ): Observable<IApplyListTemplateResult> {
    return this.sneatApiService.post<IApplyListTemplateResult>(
      'listus/list_templates_apply',
      request,
    );
  }

  public setListItemsIsCompleted(
    request: ISetListItemsIsComplete,
  ): Observable<void> {
    const url = 'listus/list_items_set_is_done';
    return this.sneatApiService.post(url, request);
  }

  public saveListItemDateTask(
    request: ISaveListItemDateTaskRequest,
  ): Observable<ISaveListItemDateTaskResponse> {
    return this.sneatApiService.post('listus/item_date_task_save', request);
  }

  public deleteListItems(request: IDeleteListItemsRequest): Observable<void> {
    return this.sneatApiService.delete(
      `listus/list_items_delete?`,
      new HttpParams({
        fromObject: {
          spaceID: request.spaceID,
          // listType: request.listType,
          listID: request.listID,
        },
      }),
      { itemIDs: request.itemIDs },
    );
  }

  public getListById(
    space: ISpaceContext,
    listType: ListType,
    listID: string,
  ): Observable<IListContext> {
    const id = this.getFullListID(listType, listID);
    const listDocRef = this.listDocRef(space.id, id);
    return this.sfs.getByDocRef(listDocRef).pipe(
      map((listContext) => {
        return this.onListSnapshot(space, listID, listType, listContext.dbo);
      }),
    );
  }

  /**
   * Optional overview capability. A missing space id or dbo is valid during a
   * cold load and yields the embedded groups (often none). When the space id is
   * known, canonical extension-owned briefs are merged with any legacy groups.
   */
  public watchListGroups(
    space: ISpaceContext,
  ): Observable<readonly IListGroup[]> {
    const embedded = embeddedListGroups(space);
    if (!space.id) {
      return of(embedded);
    }
    return this.observeSpaceLists(space.id).pipe(
      map((briefs) => mergeListGroups(embedded, listGroupsFromBriefs(briefs))),
    );
  }

  // Read-only TMDB proxy - no spaceID needed (see backend/api4listus/http_movie_search.go).
  public searchMovies(
    request: SearchMoviesRequest,
  ): Observable<SearchMoviesResponse> {
    return this.sneatApiService.post('listus/movies/search', request);
  }

  // Read-only TMDB proxy - no spaceID needed (see backend/api4listus/http_movie_resolve.go).
  public resolveMovie(
    request: ResolveMovieRequest,
  ): Observable<ResolveMovieResponse> {
    return this.sneatApiService.post('listus/movies/resolve', request);
  }

  // Identifies movies from a vague natural-language description via AI title
  // guesses grounded through TMDB search. Read-only proxy like search/resolve
  // - no spaceID needed (see backend/api4listus/http_movie_identify.go).
  public identifyMovies(
    request: IdentifyMoviesRequest,
  ): Observable<IdentifyMoviesResponse> {
    return this.sneatApiService.post('listus/movies/identify', request);
  }

  // Resolves the movie server-side (by tmdbID or query) and appends it, fully
  // enriched, to the watch list given by request.listID - or to the space's
  // canonical watch!movies list when listID is omitted, auto-creating that
  // standard list on first use (see backend/facade4listus/movie_watch.go).
  public addMovieToWatchlist(
    request: AddMovieToWatchlistRequest,
  ): Observable<AddMovieToWatchlistResponse> {
    return this.sneatApiService.post('listus/movies/add_to_watchlist', request);
  }

  public setListItemWatchWith(
    request: SetListItemWatchWithRequest,
  ): Observable<void> {
    const params = new HttpParams({
      fromObject: {
        spaceID: request.spaceID,
        listID: request.listID,
      },
    });
    return this.sneatApiService.post(
      `listus/list_items_set_watch_with?${params.toString()}`,
      { item: request.item, watchWith: request.watchWith },
    );
  }

  private readonly onListSnapshot = (
    space: ISpaceContext,
    id: string,
    type: ListType,
    dbo?: IListDbo | null,
  ): IListContext => ({
    id,
    type,
    dbo,
    brief: dbo,
    space,
  });

  private listDocRef(
    spaceID: string,
    listID: string,
  ): DocumentReference<IListDbo> {
    const listsCollection = collection(
      this.afs,
      'spaces',
      spaceID,
      'ext',
      'listus',
      'lists',
    );
    return doc(listsCollection, listID) as DocumentReference<IListDbo>;
  }
}

function embeddedListGroups(space: ISpaceContext): IListGroup[] {
  return [
    ...((space.dbo as unknown as IListusSpaceDbo | undefined)?.listGroups ??
      []),
  ];
}

/** Same qualified-id grouping the lists UI uses for extension-owned briefs. */
function listGroupsFromBriefs(
  briefs: Readonly<Record<string, IListBrief>>,
): IListGroup[] {
  const groups = new Map<ListType, IListGroup>();
  Object.entries(briefs)
    .sort(([, a], [, b]) => (a.title || '').localeCompare(b.title || ''))
    .forEach(([qualifiedID, brief]) => {
      const separator = qualifiedID.indexOf('!');
      const type = brief.type;
      if (!type || separator <= 0 || qualifiedID.slice(0, separator) !== type) {
        return;
      }
      const shortID = qualifiedID.slice(separator + 1);
      if (!shortID) return;
      let group = groups.get(type);
      if (!group) {
        group = { id: type, type, lists: [] };
        groups.set(type, group);
      }
      group.lists?.push({
        ...brief,
        id: shortID,
        shortId: shortID,
        type,
      });
    });
  return [...groups.values()];
}

function mergeListGroups(
  base: readonly IListGroup[],
  extra: readonly IListGroup[],
): IListGroup[] {
  const groups = base.map((group) => ({
    ...group,
    lists: [...(group.lists ?? [])],
  }));
  extra.forEach((incoming) => {
    const index = groups.findIndex((current) => current.type === incoming.type);
    if (index < 0) {
      groups.push({ ...incoming, lists: [...(incoming.lists ?? [])] });
      return;
    }
    const lists = groups[index].lists ?? [];
    (incoming.lists ?? []).forEach((list) => {
      if (!lists.some((current) => current.id === list.id)) {
        lists.push({ ...list });
      }
    });
    groups[index] = { ...groups[index], lists };
  });
  return groups;
}
