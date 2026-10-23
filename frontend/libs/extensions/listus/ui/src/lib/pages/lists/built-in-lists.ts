import { SpaceType } from '@sneat/core';
import {
  IListBrief,
  IListGroup,
  ListType,
} from '@sneat/extension-listus-contract';

// Built-in default list groups shown for a space, for instant UX before/alongside
// the lists persisted on the space DBO. Personal and family spaces both
// get To Buy / To Do; other space types have no built-ins. Shared by the lists
// page and the listus space menu so they stay consistent.
export function builtInListGroups(spaceType?: SpaceType): IListGroup[] {
  if (spaceType !== 'family' && spaceType !== 'personal') {
    return [];
  }
  return [
    {
      id: 'buy',
      type: 'buy',
      title: 'To Buy',
      lists: [
        { id: 'groceries', type: 'buy', emoji: '🛒', title: 'Groceries' },
        { id: 'wholesale', type: 'buy', emoji: '🛒', title: 'Wholesale' },
      ],
    },
    {
      id: 'to-do',
      type: 'do',
      title: 'To Do',
      lists: [{ id: 'chores', type: 'do', emoji: '🧹', title: 'Chores' }],
    },
    {
      id: 'watch',
      type: 'watch',
      title: 'To Watch',
      lists: [{ id: 'movies', type: 'watch', emoji: '📽️', title: 'Movies' }],
    },
  ];
}

/** Converts the canonical extension-owned list summary into display groups. */
export function listGroupsFromBriefs(
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
