import { Provider } from '@angular/core';
import { LISTUS_SERVICE } from '@sneat/extension-listus-contract';
import { ListService } from './services';

// Registers the concrete ListService and binds it to the LISTUS_SERVICE token so
// consumers depend only on the IListusService contract. Wired in at app bootstrap
// (consumers do not import this factory directly).
export function provideListusInternal(): Provider[] {
  return [ListService, { provide: LISTUS_SERVICE, useExisting: ListService }];
}
