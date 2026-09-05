import { TestBed } from '@angular/core/testing';
import { Firestore } from 'firebase/firestore';
import {
  listInfosFromSpaceDbo,
  ListusSpaceService,
} from './listus-space.service';

vi.mock('firebase/firestore');

describe('ListusSpaceService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        ListusSpaceService,
        {
          provide: Firestore,
          useValue: { type: 'Firestore', toJSON: () => ({}) },
        },
      ],
    });
  });

  it('is injectable', () => {
    expect(TestBed.inject(ListusSpaceService)).toBeTruthy();
  });

  it('projects the persisted lists map into editor-ready list info', () => {
    expect(
      listInfosFromSpaceDbo({
        lists: {
          'do!cleaning': {
            type: 'do',
            title: 'Saturday cleaning',
            createdAt: { seconds: 1_788_609_600, nanoseconds: 0 },
            createdBy: 'user-1',
          },
        },
      }),
    ).toEqual([
      expect.objectContaining({
        id: 'do!cleaning',
        shortId: 'cleaning',
        type: 'do',
        title: 'Saturday cleaning',
      }),
    ]);
  });

  it('merges legacy list groups while the canonical lists map wins duplicates', () => {
    expect(
      listInfosFromSpaceDbo({
        lists: {
          'do!cleaning': {
            type: 'do',
            title: 'Saturday cleaning',
            createdAt: { seconds: 1_788_609_600, nanoseconds: 0 },
            createdBy: 'user-1',
          },
        },
        listGroups: [
          {
            id: 'chores',
            lists: [
              {
                type: 'do',
                shortId: 'cleaning',
                title: 'Stale cleaning title',
              },
              { type: 'buy', shortId: 'groceries', title: 'Groceries' },
              { type: 'do', title: 'Built-in tasks' },
            ],
          },
        ],
      }),
    ).toEqual([
      expect.objectContaining({
        id: 'do!cleaning',
        title: 'Saturday cleaning',
      }),
      expect.objectContaining({
        shortId: 'groceries',
        title: 'Groceries',
      }),
      expect.objectContaining({ type: 'do', title: 'Built-in tasks' }),
    ]);
  });
});
