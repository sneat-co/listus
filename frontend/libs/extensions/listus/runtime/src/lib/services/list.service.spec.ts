import { TestBed } from '@angular/core/testing';
import { SneatApiService } from '@sneat/api';
import { IListBrief, IListGroup } from '@sneat/extension-listus-contract';
import { ISpaceContext } from '@sneat/space-models';
import { Firestore } from 'firebase/firestore';
import { of } from 'rxjs';
import { ListService } from './list.service';

vi.mock('firebase/firestore');

describe('ListService', () => {
  beforeEach(() => {
    TestBed.resetTestingModule();
  });

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

    service.getListById({ id: 'household-1' }, 'buy', 'groceries').subscribe();

    expect(listDocRef).toHaveBeenCalledWith('household-1', 'buy!groceries');
    expect(getByDocRef).toHaveBeenCalledWith(docRef);
  });

  it('returns embedded groups without reading extension storage when the space id is missing', () => {
    const service = createService();
    const observed = vi.spyOn(service, 'observeSpaceLists');
    const embedded: IListGroup[] = [
      {
        id: 'buy',
        type: 'buy',
        lists: [{ id: 'groceries', type: 'buy', title: 'Groceries' }],
      },
    ];
    let groups: readonly IListGroup[] | undefined;
    service
      .watchListGroups({
        dbo: { listGroups: embedded },
      } as unknown as ISpaceContext)
      .subscribe((value) => {
        groups = value;
      });

    expect(groups).toEqual(embedded);
    expect(observed).not.toHaveBeenCalled();
  });

  it('merges canonical extension briefs with legacy groups and drops malformed ids', () => {
    const service = createService();
    const created = {
      createdAt: { seconds: 1, nanoseconds: 0 },
      createdBy: 'u1',
    };
    vi.spyOn(service, 'observeSpaceLists').mockReturnValue(
      of({
        'do!payments': {
          ...created,
          type: 'do',
          title: 'Payments',
        } as IListBrief,
        'do!chores': { ...created, type: 'do', title: 'Chores' } as IListBrief,
        missing: {
          ...created,
          type: 'do',
          title: 'Missing separator',
        } as IListBrief,
        'buy!wrong': {
          ...created,
          type: 'do',
          title: 'Wrong type',
        } as IListBrief,
        'do!': {
          ...created,
          type: 'do',
          title: 'Empty short id',
        } as IListBrief,
      }),
    );
    let groups: readonly IListGroup[] | undefined;
    service
      .watchListGroups({
        id: 'household-1',
        dbo: {
          listGroups: [
            {
              id: 'buy',
              type: 'buy',
              lists: [{ id: 'groceries', type: 'buy', title: 'Groceries' }],
            },
          ],
        },
      } as unknown as ISpaceContext)
      .subscribe((value) => {
        groups = value;
      });

    expect(groups?.map((group) => group.type)).toEqual(['buy', 'do']);
    expect(groups?.[0].lists?.map((list) => list.id)).toEqual(['groceries']);
    expect(groups?.[1].lists?.map((list) => list.id)).toEqual([
      'chores',
      'payments',
    ]);
  });
});

function createService(): ListService {
  TestBed.configureTestingModule({
    providers: [
      ListService,
      { provide: Firestore, useValue: {} },
      { provide: SneatApiService, useValue: {} },
    ],
  });
  return TestBed.inject(ListService);
}
