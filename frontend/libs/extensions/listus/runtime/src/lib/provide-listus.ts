import { Provider } from '@angular/core';
import { LISTUS_SERVICE } from '@sneat/extension-listus-contract';
import { ListService, ListusSpaceService } from './services';

// Registers the concrete ListService and binds it to the LISTUS_SERVICE token so
// consumers depend only on the IListusService contract. Wired in at app bootstrap
// (consumers do not import this factory directly).
export function provideListus(): Provider[] {
  return [
    ListService,
    ListusSpaceService,
    { provide: LISTUS_SERVICE, useExisting: ListService },
  ];
}
