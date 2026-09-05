import { TestBed } from '@angular/core/testing';
import { SneatApiService } from '@sneat/api';
import { Firestore } from 'firebase/firestore';
import { of } from 'rxjs';
import { ListService } from './list.service';

vi.mock('firebase/firestore');

describe('ListService', () => {
  it('reads a short list ID from the canonical typed document ID', () => {
    TestBed.configureTestingModule({
      providers: [
        ListService,
        { provide: Firestore, useValue: {} },
        { provide: SneatApiService, useValue: {} },
      ],
    });
    const service = TestBed.inject(ListService);
    const docRef = {};
    const listDocRef = vi.fn(() => docRef);
    const getByDocRef = vi.fn(() =>
      of({ dbo: { type: 'buy', title: 'Groceries', items: [] } }),
    );
    Object.assign(service as unknown as Record<string, unknown>, {
      listDocRef,
      sfs: { getByDocRef },
    });

    service
      .getListById({ id: 'household-1' }, 'buy', 'groceries')
      .subscribe();

    expect(listDocRef).toHaveBeenCalledWith('household-1', 'buy!groceries');
    expect(getByDocRef).toHaveBeenCalledWith(docRef);
  });
});
