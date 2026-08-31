import { signal, WritableSignal } from '@angular/core';
import { IListContext } from '@sneat/extension-listus-contract';
import { BaseListPage } from './base-list-page';

describe('BaseListPage', () => {
  it('preserves the route-seeded title when a refresh provides an empty title', () => {
    const page = Object.create(BaseListPage.prototype) as {
      $list: WritableSignal<IListContext | undefined>;
      setList(list: IListContext): void;
    };
    page.$list = signal({
      id: 'groceries',
      type: 'buy',
      brief: { type: 'buy', title: 'Groceries' },
    } as IListContext);

    page.setList({
      id: 'groceries',
      type: 'buy',
      brief: { type: 'buy', title: '' },
    } as IListContext);

    expect(page.$list()?.brief?.title).toBe('Groceries');
  });

  it('preserves the title when the loaded context prefixes the route ID with its type', () => {
    const page = Object.create(BaseListPage.prototype) as {
      $list: WritableSignal<IListContext | undefined>;
      setList(list: IListContext): void;
    };
    page.$list = signal({
      id: 'groceries',
      type: 'buy',
      brief: { type: 'buy', title: 'Groceries' },
    } as IListContext);

    page.setList({
      id: 'buy!groceries',
      type: 'buy',
      brief: { type: 'buy', title: '' },
    } as IListContext);

    expect(page.$list()?.brief?.title).toBe('Groceries');
  });
});
