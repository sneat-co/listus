import { signal, WritableSignal } from '@angular/core';
import { IListContext } from '@sneat/extension-listus-contract';
import { ISpaceContext } from '@sneat/space-models';
import { Observable, of, Subject } from 'rxjs';
import { vi } from 'vitest';
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

  describe('deleteAll() / deleteCompleted()', () => {
    type DeletablePage = {
      allListItems: WritableSignal<IListItemWithUiState[] | undefined>;
      deleteItems(items?: IListItemWithUiState[]): void;
      deleteAll(): void;
      deleteCompleted(): void;
    };

    const items = [
      { brief: { id: 'a', status: 'active' }, state: {} },
      { brief: { id: 'b', status: 'done' }, state: {} },
    ] as IListItemWithUiState[];

    const setup = () => {
      const page = Object.create(
        ListPageComponent.prototype,
      ) as DeletablePage;
      page.allListItems = signal(items);
      const deleteItemsSpy = vi.fn();
      (page as unknown as { deleteItems: unknown }).deleteItems =
        deleteItemsSpy;
      return { page, deleteItemsSpy };
    };

    afterEach(() => {
      vi.restoreAllMocks();
    });

    it('deletes everything only after the user confirms', () => {
      const { page, deleteItemsSpy } = setup();
      const confirmSpy = vi
        .spyOn(globalThis, 'confirm')
        .mockReturnValueOnce(false);

      page.deleteAll();
      expect(confirmSpy).toHaveBeenCalledOnce();
      expect(deleteItemsSpy).not.toHaveBeenCalled();

      confirmSpy.mockReturnValueOnce(true);
      page.deleteAll();
      expect(deleteItemsSpy).toHaveBeenCalledExactlyOnceWith(items);
    });

    it('deletes only completed items, and only after the user confirms', () => {
      const { page, deleteItemsSpy } = setup();
      vi.spyOn(globalThis, 'confirm').mockReturnValue(true);

      page.deleteCompleted();

      expect(deleteItemsSpy).toHaveBeenCalledExactlyOnceWith([items[1]]);
    });

    it('skips the confirm and the delete when there is nothing to delete', () => {
      const page = Object.create(
        ListPageComponent.prototype,
      ) as DeletablePage;
      page.allListItems = signal([]);
      const deleteItemsSpy = vi.fn();
      (page as unknown as { deleteItems: unknown }).deleteItems =
        deleteItemsSpy;
      const confirmSpy = vi.spyOn(globalThis, 'confirm');
      vi.spyOn(globalThis, 'alert').mockImplementation(() => undefined);

      page.deleteAll();

      expect(confirmSpy).not.toHaveBeenCalled();
      expect(deleteItemsSpy).not.toHaveBeenCalled();
    });
  });

  it('navigates to the built-in Groceries list', () => {
    const navigateForwardToSpacePage = vi.fn().mockResolvedValue(true);
    const space = { id: 'family' } as ISpaceContext;
    const page = Object.create(ListPageComponent.prototype) as {
      space: ISpaceContext;
      spaceNav: {
        navigateForwardToSpacePage: typeof navigateForwardToSpacePage;
      };
      errorLogger: { logErrorHandler: () => (error: unknown) => void };
      goGroceries(): void;
    };
    // `space` and `spaceNav` are getter-only accessors on the SpaceBaseComponent
    // prototype (no setter), so a plain assignment would throw - shadow them
    // with own properties instead.
    Object.defineProperty(page, 'space', { value: space });
    Object.defineProperty(page, 'spaceNav', {
      value: { navigateForwardToSpacePage },
    });
    page.errorLogger = { logErrorHandler: () => () => undefined };

    page.goGroceries();

    expect(navigateForwardToSpacePage).toHaveBeenCalledExactlyOnceWith(
      space,
      'list/buy/groceries',
    );
  });
});
