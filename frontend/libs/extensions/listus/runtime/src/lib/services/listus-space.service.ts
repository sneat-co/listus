import { Injectable, inject, Injector } from '@angular/core';
import {
  IListInfo,
  IListusSpaceDbo,
} from '@sneat/extension-listus-contract';
import { SpaceModuleService } from '@sneat/space-services';
import { Firestore } from 'firebase/firestore';
import { map, Observable } from 'rxjs';
import { listShortID } from '../list-id';

export function listInfosFromSpaceDbo(
  dbo?: IListusSpaceDbo | null,
): readonly IListInfo[] {
  const persisted = Object.entries(dbo?.lists ?? {}).map(([id, brief]) => ({
    ...brief,
    id,
    shortId: listShortID(id, brief.type),
  }));
  const persistedIDs = new Set(persisted.map((list) => list.id));
  const legacy = (dbo?.listGroups ?? [])
    .flatMap((group) => group.lists ?? [])
    .filter((list) => {
      const fullID =
        list.id ??
        (list.shortId && list.type ? `${list.type}!${list.shortId}` : undefined);
      return !fullID || !persistedIDs.has(fullID);
    });
  return [...persisted, ...legacy];
}

/** Watches the Listus module record at spaces/{spaceID}/ext/listus. */
@Injectable()
export class ListusSpaceService extends SpaceModuleService<IListusSpaceDbo> {
  public constructor() {
    super(inject(Injector), 'listus', inject(Firestore));
  }

  public watchListInfos(spaceID: string): Observable<readonly IListInfo[]> {
    return this.watchSpaceModuleRecord(spaceID).pipe(
      map((record) => listInfosFromSpaceDbo(record.dbo)),
    );
  }
}
