import { signal, WritableSignal } from '@angular/core';
import { IListContext } from '@sneat/extension-listus-contract';
import { Observable, of, Subject } from 'rxjs';
import { IListItemWithUiState } from './list-item-with-ui-state';
import { ListPageComponent } from './list-page.component';

describe('ListPageComponent', () => {
  it('keeps a completing active item rendered until the save resolves', () => {
    const oldItem = {
      brief: { id: 'milk', status: undefined },
      state: {},
    } as IListItemWithUiState;
    const savingItem = {
      brief: { ...oldItem.brief, status: 'done' },
      state: { isChangingIsDone: true },
    } as IListItemWithUiState;
    const savedItem = {
      brief: savingItem.brief,
      state: { isChangingIsDone: false },
    } as IListItemWithUiState;
    const page = Object.create(ListPageComponent.prototype) as {
      allListItems: WritableSignal<IListItemWithUiState[] | undefined>;
      listItems: WritableSignal<IListItemWithUiState[] | undefined>;
      $listType: WritableSignal<undefined>;
      doneFilter: WritableSignal<'all' | 'active' | 'completed' | undefined>;
      isHideWatched: WritableSignal<boolean>;
      addingItems: IListItemWithUiState[];
      itemChanged(changedItem: {
        old: IListItemWithUiState;
        new: IListItemWithUiState;
      }): void;
    };
    page.allListItems = signal([oldItem]);
    page.listItems = signal([oldItem]);
    page.$listType = signal(undefined);
    page.doneFilter = signal('active');
    page.isHideWatched = signal(false);
    page.addingItems = [];

    page.itemChanged({ old: oldItem, new: savingItem });

    expect(page.listItems()).toEqual([savingItem]);

    page.itemChanged({ old: savingItem, new: savedItem });

    expect(page.listItems()).toEqual([]);
  });

  it('switches to Active after reactivating completed items', () => {
    const page = Object.create(ListPageComponent.prototype) as {
      params: {
        listService: {
          setListItemsIsCompleted: () => Observable<void>;
        };
      };
      $list: WritableSignal<IListContext | undefined>;
      $space: WritableSignal<{ id: string }>;
      $listType: WritableSignal<undefined>;
      destroyed$: Subject<void>;
      doneFilter: WritableSignal<'all' | 'active' | 'completed' | undefined>;
      allListItems: WritableSignal<IListItemWithUiState[] | undefined>;
      listItems: WritableSignal<IListItemWithUiState[] | undefined>;
      isHideWatched: WritableSignal<boolean>;
      performing: WritableSignal<'reactivating completed' | undefined>;
      addingItems: IListItemWithUiState[];
      reactivateCompleted(): void;
    };
    page.params = {
      listService: { setListItemsIsCompleted: () => of(void 0) },
    };
    page.$list = signal({
      id: 'groceries',
      type: 'buy',
      brief: { type: 'buy', title: 'Groceries' },
    } as IListContext);
    page.$space = signal({ id: 'personal' });
    page.$listType = signal(undefined);
    page.destroyed$ = new Subject<void>();
    page.doneFilter = signal('completed');
    page.allListItems = signal([
      { brief: { id: 'milk', status: 'done' }, state: {} },
    ] as IListItemWithUiState[]);
    page.listItems = signal(undefined);
    page.isHideWatched = signal(false);
    page.performing = signal(undefined);
    page.addingItems = [];

    page.reactivateCompleted();

    expect(page.doneFilter()).toBe('active');
  });
});
