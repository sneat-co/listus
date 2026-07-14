import { LISTUS_SERVICE } from '@sneat/extension-listus-contract';
import { ListService } from './services';
import { provideListus } from './provide-listus';

describe('provideListus', () => {
  it('provides ListService and binds it to LISTUS_SERVICE', () => {
    const providers = provideListus();
    expect(providers).toContain(ListService);
    expect(providers).toContainEqual({
      provide: LISTUS_SERVICE,
      useExisting: ListService,
    });
  });
});
