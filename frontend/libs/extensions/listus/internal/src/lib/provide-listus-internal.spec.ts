import { LISTUS_SERVICE } from '@sneat/extension-listus-contract';
import { ListService } from './services';
import { provideListusInternal } from './provide-listus-internal';

describe('provideListusInternal', () => {
  it('provides ListService and binds it to LISTUS_SERVICE', () => {
    const providers = provideListusInternal();
    expect(providers).toContain(ListService);
    expect(providers).toContainEqual({
      provide: LISTUS_SERVICE,
      useExisting: ListService,
    });
  });
});
