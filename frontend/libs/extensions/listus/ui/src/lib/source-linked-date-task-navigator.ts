import { inject, Injectable } from '@angular/core';
import type {
  ISourceLinkedDateTaskMetadata,
  ISourceLinkedDateTaskNavigator,
} from '@sneat/extension-calendarius-contract';
import type { IRelatedModules } from '@sneat/dto';
import type { ISpaceContext } from '@sneat/space-models';
import { SPACE_NAV_SERVICE } from '@sneat/space-services';

export interface IListusDateTaskDestination {
  readonly listType: string;
  readonly listID: string;
  readonly itemID: string;
}

@Injectable()
export class ListusSourceLinkedDateTaskNavigator
  implements ISourceLinkedDateTaskNavigator
{
  private readonly spaceNav = inject(SPACE_NAV_SERVICE);

  navigateToSourceLinkedDateTask(
    space: ISpaceContext,
    task: ISourceLinkedDateTaskMetadata,
    related: IRelatedModules,
  ): Promise<boolean> {
    const destination = findListusDateTaskDestination(space.id, related);
    if (
      !destination ||
      task.source.ownerSpaceID !== space.id ||
      !task.actionID ||
      (task.actionDisposition !== 'navigate' &&
        task.actionDisposition !== 'requires_input')
    ) {
      return Promise.resolve(false);
    }
    return this.spaceNav.navigateForwardToSpacePage(
      space,
      `list/${destination.listType}/${destination.listID}`,
      { queryParams: { itemID: destination.itemID } },
    );
  }
}

export function findListusDateTaskDestination(
  spaceID: string,
  related: IRelatedModules,
): IListusDateTaskDestination | undefined {
  const lists = related['listus']?.['lists'] || {};
  const matches: IListusDateTaskDestination[] = [];
  for (const [qualifiedListID, relation] of Object.entries(lists)) {
    const [listID, explicitSpaceID] = splitQualifiedID(qualifiedListID);
    if (explicitSpaceID !== undefined && explicitSpaceID !== spaceID) {
      continue;
    }
    const separator = listID.indexOf('!');
    if (separator <= 0 || separator === listID.length - 1) {
      continue;
    }
    for (const [subPath, child] of Object.entries(relation.subPaths || {})) {
      if (!child.rolesOfItem?.['todo-item'] && !child.rolesToItem?.['todo-item']) {
        continue;
      }
      const itemID = parseListItemSubPath(subPath);
      if (itemID) {
        matches.push({
          listType: listID.slice(0, separator),
          listID: listID.slice(separator + 1),
          itemID,
        });
      }
    }
  }
  return matches.length === 1 ? matches[0] : undefined;
}

function splitQualifiedID(value: string): readonly [string, string?] {
  const separator = value.lastIndexOf('@');
  if (separator === value.length - 1) {
    return [value, ''];
  }
  return separator < 0
    ? [value]
    : [value.slice(0, separator), value.slice(separator + 1)];
}

function parseListItemSubPath(value: string): string | undefined {
  const prefix = '/items/@id=';
  if (!value.startsWith(prefix) || value.indexOf('/', prefix.length) >= 0) {
    return undefined;
  }
  const encoded = value.slice(prefix.length);
  if (!encoded || /~(?![01])/u.test(encoded)) {
    return undefined;
  }
  return encoded.split('~1').join('/').split('~0').join('~');
}
