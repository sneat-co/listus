import { Injectable, inject } from '@angular/core';
import {
  IListBrief,
  IListDbo,
  IListusService,
  LISTUS_SERVICE,
} from '@sneat/extension-listus-contract';
import { SpaceComponentBaseParams } from '@sneat/space-components';
import { ModuleSpaceItemService } from '@sneat/space-services';

// The listus service obtained via the contract token. BaseListPage passes it to
// the SpaceItemPageBaseComponent super constructor, which expects a concrete
// ModuleSpaceItemService<IListBrief, IListDbo>; the bound implementation extends
// exactly that, so the injected value is typed as the intersection.
export type ListusServiceWithSpaceItem = IListusService &
  ModuleSpaceItemService<IListBrief, IListDbo>;

@Injectable()
export class ListusComponentBaseParams {
  readonly spaceParams = inject(SpaceComponentBaseParams);
  readonly listService = inject<ListusServiceWithSpaceItem>(LISTUS_SERVICE);
}
