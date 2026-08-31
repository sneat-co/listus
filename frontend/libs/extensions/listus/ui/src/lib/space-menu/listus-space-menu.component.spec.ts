import { IUserSpaceBrief } from '@sneat/auth-models';
import { IIdAndBrief } from '@sneat/core';
import { ISpaceContext } from '@sneat/space-models';
import { Observable, of, Subject } from 'rxjs';
import { vi } from 'vitest';
import { ListusSpaceMenuComponent } from './listus-space-menu.component';

describe('ListusSpaceMenuComponent', () => {
  it('preserves a matching list route when switching spaces', () => {
    const navigateForwardToSpacePage = vi.fn().mockResolvedValue(true);
    const menu = Object.create(ListusSpaceMenuComponent.prototype) as {
      router: { url: string };
      spaceParams: {
        spaceNavService: {
          navigateForwardToSpacePage: typeof navigateForwardToSpacePage;
        };
      };
      errorLogger: { logErrorHandler: () => (error: unknown) => void };
      spaceService: { watchSpace: () => Observable<ISpaceContext> };
      destroyed$: Subject<void>;
      switchSpace(space: IIdAndBrief<IUserSpaceBrief>): void;
    };
    menu.router = { url: '/space/personal/current/list/buy/groceries' };
    menu.spaceParams = { spaceNavService: { navigateForwardToSpacePage } };
    menu.errorLogger = { logErrorHandler: () => () => undefined };
    menu.spaceService = { watchSpace: () => of({} as ISpaceContext) };
    menu.destroyed$ = new Subject<void>();

    menu.switchSpace({
      id: 'family',
      brief: { type: 'family', title: 'Family' },
    } as IIdAndBrief<IUserSpaceBrief>);

    expect(navigateForwardToSpacePage).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'family', type: 'family' }),
      'list/buy/groceries',
      { replaceUrl: true },
    );
  });

  it('falls back to Lists when the destination lacks the current list', () => {
    const navigateForwardToSpacePage = vi.fn().mockResolvedValue(true);
    const menu = Object.create(ListusSpaceMenuComponent.prototype) as {
      router: { url: string };
      spaceParams: {
        spaceNavService: {
          navigateForwardToSpacePage: typeof navigateForwardToSpacePage;
        };
      };
      errorLogger: { logErrorHandler: () => (error: unknown) => void };
      spaceService: { watchSpace: () => Observable<ISpaceContext> };
      destroyed$: Subject<void>;
      switchSpace(space: IIdAndBrief<IUserSpaceBrief>): void;
    };
    menu.router = { url: '/space/personal/current/list/buy/custom-list' };
    menu.spaceParams = { spaceNavService: { navigateForwardToSpacePage } };
    menu.errorLogger = { logErrorHandler: () => () => undefined };
    menu.spaceService = {
      watchSpace: () =>
        of({ id: 'family', type: 'family', dbo: null } as ISpaceContext),
    };
    menu.destroyed$ = new Subject<void>();

    menu.switchSpace({
      id: 'family',
      brief: { type: 'family', title: 'Family' },
    } as IIdAndBrief<IUserSpaceBrief>);

    expect(navigateForwardToSpacePage).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'family', type: 'family' }),
      'lists',
      { replaceUrl: true },
    );
  });
});
