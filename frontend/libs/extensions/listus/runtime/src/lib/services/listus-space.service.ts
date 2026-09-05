import { Injectable, inject, Injector } from '@angular/core';
import { IListusSpaceDbo } from '@sneat/extension-listus-contract';
import { SpaceModuleService } from '@sneat/space-services';
import { Firestore } from 'firebase/firestore';

/** Watches the Listus module record at spaces/{spaceID}/ext/listus. */
@Injectable()
export class ListusSpaceService extends SpaceModuleService<IListusSpaceDbo> {
  public constructor() {
    super(inject(Injector), 'listus', inject(Firestore));
  }
}
