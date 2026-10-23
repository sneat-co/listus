import { InjectionToken } from '@angular/core';
import { IListGroup } from '@sneat/extension-listus-contract';
import { ISpaceContext } from '@sneat/space-models';
import { Observable } from 'rxjs';

/**
 * Optional overview source. Hosts that do not provide the token keep the
 * built-in lists plus the extension-owned summary from `observeSpaceLists`.
 *
 * Defined here, not in `@sneat/extension-listus-contract`, because the published
 * contract does not export it and the UI library cannot import the runtime library.
 */
export const LISTUS_LIST_GROUPS_READER =
  new InjectionToken<IListusListGroupsReader>('LISTUS_LIST_GROUPS_READER');

export interface IListusListGroupsReader {
  watchListGroups(space: ISpaceContext): Observable<readonly IListGroup[]>;
}
