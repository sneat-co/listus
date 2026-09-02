import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  signal,
} from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { IUserSpaceBrief } from '@sneat/auth-models';
import {
  IonIcon,
  IonItem,
  IonLabel,
  IonList,
  IonNote,
  MenuController,
} from '@ionic/angular';
import { AuthMenuItemComponent } from '@sneat/auth-ui';
import { ContactusServicesModule } from '@sneat/extension-contactus';
import {
  SpaceBaseComponent,
  SpaceComponentBaseParams,
  SpaceSelectorComponent,
} from '@sneat/space-components';
import { IIdAndBrief } from '@sneat/core';
import { ISpaceContext } from '@sneat/space-models';
import { SpaceServiceModule } from '@sneat/space-services';
import { ClassName } from '@sneat/ui';
import { filter, takeUntil, take } from 'rxjs/operators';
import { Subscription } from 'rxjs';
import {
  IListGroup,
  IListusSpaceDbo,
  IListusListGroupsReader,
  LISTUS_LIST_GROUPS_READER,
  ListType,
} from '@sneat/extension-listus-contract';
import { builtInListGroups } from '../pages/lists/built-in-lists';

// listus-specific side menu rendered in the space "menu" outlet. Unlike the
// generic @sneat SpaceMenuComponent (which hardcodes every sneat-app extension —
// Assets, Budget, Calendar, Contacts, Debts, …, none of which exist in
// listus-app), this shows only what listus has: a space selector (to switch
// spaces, like sneat-app) and the selected space's lists.
@Component({
  selector: 'listus-space-menu',
  templateUrl: './listus-space-menu.component.html',
  imports: [
    RouterLink,
    ContactusServicesModule,
    SpaceServiceModule,
    IonList,
    IonItem,
    IonIcon,
    IonLabel,
    IonNote,
    AuthMenuItemComponent,
    SpaceSelectorComponent,
  ],
  providers: [
    { provide: ClassName, useValue: 'ListusSpaceMenuComponent' },
    SpaceComponentBaseParams,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ListusSpaceMenuComponent extends SpaceBaseComponent {
  protected readonly $disabled = computed(() => !this.$spaceID());
  protected readonly $listGroups = signal<IListGroup[]>([]);

  private readonly menuCtrl = inject(MenuController);
  private readonly router = inject(Router);
  private readonly listGroupsReader = inject<IListusListGroupsReader>(
    LISTUS_LIST_GROUPS_READER,
    { optional: true },
  );
  private listGroupsSubscription?: Subscription;

  constructor() {
    super();
    // Seed the built-in default lists (e.g. family To Buy / To Do) as soon as the
    // space type is known from the URL, before the space document loads. Mirrors
    // the lists page so the menu shows lists instantly; onSpaceDboChanged() below
    // re-seeds + merges persisted lists once the DBO arrives.
    this.spaceTypeChanged$
      .pipe(takeUntil(this.destroyed$))
      .subscribe((spaceType) => {
        if (spaceType && !this.$listGroups().length) {
          this.$listGroups.set([...builtInListGroups(spaceType)]);
        }
      });
  }

  // Mirror the lists page: built-in defaults (family) + the lists persisted on
  // the space DBO, deduped by group type.
  protected override onSpaceDboChanged(): void {
    super.onSpaceDboChanged();
    this.listGroupsSubscription?.unsubscribe();
    this.listGroupsSubscription = undefined;
    const groups: IListGroup[] = this.space
      ? [...builtInListGroups(this.space.type)]
      : [];
    const dbo = this.space?.dbo as unknown as IListusSpaceDbo | undefined;
    (dbo?.listGroups || []).forEach((g) => {
      if (!groups.some((x) => x.type === g.type)) {
        groups.push(g);
      }
    });
    this.$listGroups.set(groups);
    if (!this.space || !this.listGroupsReader) return;
    this.listGroupsSubscription = this.listGroupsReader
      .watchListGroups(this.space)
      .pipe(takeUntil(this.destroyed$))
      .subscribe({
        next: (storageGroups) => {
          const merged = [...builtInListGroups(this.space.type)];
          storageGroups.forEach((group) => {
            const index = merged.findIndex((current) => current.type === group.type);
            if (index >= 0) merged[index] = group;
            else merged.push(group);
          });
          this.$listGroups.set(merged);
        },
        error: (error: unknown) => {
          this.$listGroups.set([...builtInListGroups(this.space.type)]);
          this.errorLogger.logErrorHandler(
            'Failed to load storage-backed list overview',
          )(error);
        },
      });
  }

  protected closeMenu(): void {
    this.menuCtrl.close().catch(this.errorLogger.logError);
  }

  protected switchSpace(spaceRef: IIdAndBrief<IUserSpaceBrief>): void {
    const currentList = this.currentListRoute();
    const target = {
      id: spaceRef.id,
      type: spaceRef.brief.type,
      brief: spaceRef.brief,
    } as ISpaceContext;

    if (!currentList || this.hasList(target, currentList.type, currentList.id)) {
      this.navigateToSelectedSpace(target, currentList);
      return;
    }

    // Built-in lists can be determined from the space type immediately. For a
    // custom list, wait for the selected space document before deciding whether
    // its matching list route is valid.
    this.spaceService
      .watchSpace(spaceRef.id)
      .pipe(
        filter((space) => space.dbo !== undefined),
        take(1),
        takeUntil(this.destroyed$),
      )
      .subscribe({
        next: (space) => this.navigateToSelectedSpace(space, currentList),
        error: this.errorLogger.logErrorHandler(
          'Failed to load selected space before navigating',
        ),
      });
  }

  private navigateToSelectedSpace(
    space: ISpaceContext,
    currentList?: { type: ListType; id: string },
  ): void {
    const page =
      currentList && this.hasList(space, currentList.type, currentList.id)
        ? `list/${currentList.type}/${currentList.id}`
        : 'lists';
    this.spaceNav
      .navigateForwardToSpacePage(space, page, { replaceUrl: true })
      .catch(
        this.errorLogger.logErrorHandler('Failed to navigate to selected space'),
      );
  }

  private currentListRoute(): { type: ListType; id: string } | undefined {
    // The menu can coexist with any Listus content route. Read the URL rather
    // than navigation state so the same behavior works after a deep-link refresh.
    const routeMatch = this.router.url.match(
      /^\/space\/[^/]+\/[^/]+\/list\/([^/?#]+)\/([^/?#]+)/,
    );
    if (!routeMatch) {
      return undefined;
    }
    return { type: routeMatch[1] as ListType, id: routeMatch[2] };
  }

  private hasList(space: ISpaceContext, type: ListType, id: string): boolean {
    const persistedGroups = (space.dbo as IListusSpaceDbo | undefined)
      ?.listGroups;
    return [...builtInListGroups(space.type), ...(persistedGroups || [])].some(
      (group) =>
        group.lists?.some(
          (list) => list.type === type && list.id === id,
        ),
    );
  }
}
