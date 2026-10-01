import { Provider } from '@angular/core';
import { ListusComponentBaseParams } from './listus-component-base-params';
import { ListDialogsService } from './pages/dialogs/ListDialogs.service';

/**
 * What the exported `listus-list-item` and `listus-new-list-item` components inject
 * besides the host's own space services. The listus routes provide these for the
 * list pages; a host that embeds the components elsewhere (e.g. the Sneat chat) adds
 * them to the embedding component's `providers`.
 */
export const LISTUS_LIST_ITEM_PROVIDERS: Provider[] = [
  ListusComponentBaseParams,
  ListDialogsService,
];
