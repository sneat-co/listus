import { signal, WritableSignal } from '@angular/core';
import { IListGroup } from '@sneat/extension-listus-contract';
import { SpaceBaseComponent } from '@sneat/space-components';
import { ISpaceContext } from '@sneat/space-models';
import { Subject } from 'rxjs';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ListsPageComponent } from './pages/lists/lists-page.component';
import { ListusSpaceMenuComponent } from './space-menu/listus-space-menu.component';

type OverviewHarness = {
  space: ISpaceContext | undefined;
  $listGroups: WritableSignal<IListGroup[] | undefined>;
  listGroupsReader?: { watchListGroups: ReturnType<typeof vi.fn> };
  destroyed$: Subject<void>;
  errorLogger: {
    logErrorHandler: ReturnType<typeof vi.fn>;
    logError: ReturnType<typeof vi.fn>;
  };
  onSpaceDboChanged(): void;
};

function group(...ids: string[]): IListGroup {
  return {
    id: 'do', type: 'do', title: 'To do',
    lists: ids.map(id => ({ id, type: 'do', title: id })),
  } as IListGroup;
}

describe.each([
  ['lists page', ListsPageComponent.prototype],
  ['space menu', ListusSpaceMenuComponent.prototype],
])('%s optional overview reader', (_name, prototype) => {
  beforeEach(() => {
    vi.spyOn(
      SpaceBaseComponent.prototype as unknown as { onSpaceDboChanged(): void },
      'onSpaceDboChanged',
    ).mockImplementation(() => undefined);
  });
  afterEach(() => vi.restoreAllMocks());

  function fixture() {
    const updates = new Subject<readonly IListGroup[]>();
    const failure = vi.fn();
    const view = Object.create(prototype) as OverviewHarness;
    Object.defineProperty(view, 'space', {
      value: { id: 'one', type: 'group' } as ISpaceContext,
      writable: true,
    });
    view.$listGroups = signal<IListGroup[] | undefined>([]);
    view.destroyed$ = new Subject<void>();
    view.listGroupsReader = { watchListGroups: vi.fn(() => updates) };
    view.errorLogger = { logErrorHandler: vi.fn(() => failure), logError: vi.fn() };
    return { view, updates, failure };
  }

  it('replaces snapshots so removed lists disappear', () => {
    const { view, updates } = fixture();
    view.onSpaceDboChanged();
    updates.next([group('first', 'removed')]);
    expect(view.$listGroups()?.[0].lists).toHaveLength(2);
    updates.next([group('first')]);
    expect(view.$listGroups()?.[0].lists?.map(list => list.id)).toEqual(['first']);
    updates.next([]);
    expect(view.$listGroups()).toEqual([]);
  });

  it('cancels the old Space on a null transition without reading an absent Space', () => {
    const { view, updates } = fixture();
    view.onSpaceDboChanged();
    expect(updates.observed).toBe(true);
    view.space = undefined;
    view.onSpaceDboChanged();
    expect(updates.observed).toBe(false);
    expect(view.listGroupsReader?.watchListGroups).toHaveBeenCalledTimes(1);
    updates.next([group('stale')]);
    expect(view.$listGroups()).toEqual([]);
  });

  it('cancels stale subscriptions when switching Spaces and on destruction', () => {
    const { view, updates } = fixture();
    const second = new Subject<readonly IListGroup[]>();
    view.onSpaceDboChanged();
    view.space = { id: 'two', type: 'group' } as ISpaceContext;
    view.listGroupsReader!.watchListGroups.mockReturnValue(second);
    view.onSpaceDboChanged();
    expect(updates.observed).toBe(false);
    second.next([group('current')]);
    updates.next([group('stale')]);
    expect(view.$listGroups()?.[0].lists?.[0].id).toBe('current');
    view.destroyed$.next();
    expect(second.observed).toBe(false);
  });

  it('clears failed storage snapshots and reports the error', () => {
    const { view, updates, failure } = fixture();
    view.onSpaceDboChanged();
    updates.next([group('stale')]);
    const error = new Error('session expired');
    updates.error(error);
    expect(failure).toHaveBeenCalledWith(error);
    expect(view.$listGroups()).toEqual([]);
  });

  it('retains normal family built-ins when no reader is bound', () => {
    const { view } = fixture();
    view.listGroupsReader = undefined;
    view.space = { id: 'family', type: 'family' } as ISpaceContext;
    view.onSpaceDboChanged();
    expect(view.$listGroups()?.map(item => item.type)).toEqual(expect.arrayContaining(['do', 'buy']));
  });
});
