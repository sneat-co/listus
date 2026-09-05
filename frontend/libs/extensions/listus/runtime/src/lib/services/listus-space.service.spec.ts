import { TestBed } from '@angular/core/testing';
import { Firestore } from 'firebase/firestore';
import { ListusSpaceService } from './listus-space.service';

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
});
