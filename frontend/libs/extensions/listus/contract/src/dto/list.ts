import { IRecord } from '@sneat/data';
import {
  IShortSpaceInfo,
  IWithCreated,
  IWithRestrictions,
  IWithSpaceIDs,
  SneatRecordStatus,
} from '@sneat/dto';

export type ListStatus = SneatRecordStatus;

export interface IQuantity {
  value: number;
  unit: string;
}

export interface IListItemCommon extends IListCommon {
  subListId?: string;
  subListType?: ListType;
  quantity?: IQuantity;
  category?: string;
}

export type IListItemBase = IListItemCommon;

export type ListItemStatus = 'done' | 'active';

// IWatchWith mirrors the Go WatchWith struct (dbo4listus/listitem.go) - who a
// watch-list movie is (or was) watched with. Denormalized id + display title,
// no join needed at render time.
export type WatchWithMode = 'alone' | 'space' | 'contact';

export interface IWatchWith {
  mode: WatchWithMode;
  // spaceID (mode==='space') or contactID (mode==='contact'); empty for 'alone'.
  ref?: string;
  // Denormalized display name (e.g. space title or contact name).
  title?: string;
}

// Movie fields on a watch-typed list item. All optional so non-watch lists
// (buy/do/cook/etc.) are completely unaffected. Mirrors the Go
// ListItemBase movie fields 1:1 (dbo4listus/listitem.go).
export interface IWatchMovieFields {
  tmdbID?: number;
  year?: number;
  posterURL?: string;
  overview?: string;
  trailerYouTubeKey?: string;
  cast?: string[]; // top ~5 cast member names, denormalized
  watchWith?: IWatchWith;
}

export interface IListItemBrief extends IListItemBase, IWatchMovieFields {
  id: string;
  readonly created?: string; // UTC datetime
  readonly emoji?: string;
  readonly status?: ListItemStatus;
  readonly img?: string;
}

// Convenience view type for rendering a watch-list item as a movie card.
export type WatchMovieItem = IListItemBrief &
  Required<Pick<IWatchMovieFields, 'tmdbID'>>;

export interface ListCounts {
  // TODO: Use some enumerator as IDB library does.
  active?: number;
  completed?: number;
}

// Kept in sync with the Go IsKnownListType() whitelist in
// backend/dbo4listus/list_dbo.go - reconciled per audit finding RM-5
// (frontend previously had cook/other/recipes/rsvp missing from Go; Go had
// general/read missing from the frontend union).
export type ListType =
  | 'general'
  | 'buy'
  | 'watch'
  | 'cook'
  | 'do'
  | 'other'
  | 'recipes'
  | 'read'
  | 'rsvp';

// IListCommon is a common base class for a List & ListItem
export interface IListCommon {
  // Do not extend from IWithCreated as it is not applicable for ICreateListItemRequest
  title: string;
  img?: string;
  emoji?: string;
  isDone?: boolean;
}

export interface IListBase extends IListCommon, IWithSpaceIDs {
  type: ListType;
  shortId?: string;
  status?: ListStatus;
}

export interface IListDbo extends IListBase, IWithRestrictions, IWithCreated {
  dtClosed?: number;
  note?: string; // Is used for example for recipe text
  numberOf?: ListCounts;
  items?: IListItemBrief[];
  commune?: IShortSpaceInfo; // Used just for in-memory columns?
}

export class ListItemInfoModel {
  static trackBy: (
    index: number,
    item: IListItemBrief,
  ) => string | number | undefined = (index, item) =>
    !item
      ? index
      : (!!item.id && `id:${item.id}`) ||
        (item.subListId && `subList:${item.subListId}`) ||
        item.title;
}

export class ListItemModel {
  static equalListItems(...items: IListItemBrief[]): boolean {
    const { id, title, subListId, category, subListType } = items[0];
    return !items.some((item) => {
      if (id) {
        return item.id !== id;
      }
      return (
        (!!title && item.title !== title) ||
        (!!subListId && item.subListId !== subListId) ||
        (!!category && item.category !== category) ||
        (!!subListType && item.subListType !== subListType)
      );
    });
  }
}

export interface IListItemDbo extends IListBase, IListItemCommon {
  listId?: string;
  score?: number;
  subListItems?: IListItemBrief[];
}

export function getListShortUrlId(
  communeId: string,
  shortId?: string,
  id?: string,
): string | undefined {
  if (shortId) {
    return `${communeId}-${shortId}`;
  }
  if (id) {
    return id;
  }
  return undefined;
}

export interface IListInfo extends IWithRestrictions {
  parentListId?: string;
  parentListType?: ListType;
  type: ListType;
  id?: string;
  shortId?: string;
  title?: string;
  hidden?: boolean;
  space?: IShortSpaceInfo;
  emoji?: string;
  img?: string;
  note?: string;
  itemsCount?: number;
}

export interface IListBrief extends IListBase, IWithCreated {
  emoji?: string;
}

export function isListInfoMatchesListDto(
  i: IListInfo,
  l: IRecord<IListDbo>,
): boolean {
  return (
    (!!i.id && i.id === l.id) ||
    (i.type === l.dbo?.type && !!i.shortId && i.shortId === l.dbo?.shortId)
  );
}

export function createListInfoFromDto(
  dto: IListDbo,
  shortId?: string,
): IListInfo {
  if (!dto.title) {
    throw new Error('!title');
  }
  const listInfo: IListInfo = {
    type: dto.type,
    title: dto.title,
  };
  if (shortId) {
    listInfo.shortId = shortId;
  }
  if (dto.items && dto.items.length) {
    listInfo.itemsCount = dto.items.length;
  }
  if (dto.emoji) {
    listInfo.emoji = dto.emoji;
  }
  if (dto.restrictions) {
    listInfo.restrictions = dto.restrictions;
  }
  if (dto.commune) {
    listInfo.space = dto.commune;
  }
  return listInfo;
}

// export function createListItemInfoFromListInfo(listInfo: IListInfo): IListItemBrief {
// 	return {
// 		id: listInfo.id || '',
// 		title: listInfo.title || '',
// 		subListType: listInfo.type,
// 		subListId: listInfo.id || `${listInfo.team && listInfo.team.id}-${listInfo.shortId}`,
// 		emoji: listInfo.emoji,
// 		img: listInfo.img,
// 	};
// }

// export function createListItemInfo(listItem: IListItemDto): IListItemBrief {
// 	const v: IListItemBrief = {
// 		id: listItem.id,
// 		title: listItem.title,
// 	};
// 	if (listItem.emoji) {
// 		v.emoji = listItem.emoji;
// 	}
// 	if (listItem.done) {
// 		v.done = true;
// 	}
// 	return v;
// }
