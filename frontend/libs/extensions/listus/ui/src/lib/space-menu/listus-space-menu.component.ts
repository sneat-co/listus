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
import { switchMap, takeUntil, take, tap } from 'rxjs/operators';
import { of, Subscription } from 'rxjs';
import {
  IListGroup,
  IListusService,
  IListusSpaceDbo,
  LISTUS_SERVICE,
  ListType,
} from '@sneat/extension-listus-contract';
import {
  IListusListGroupsReader,
  LISTUS_LIST_GROUPS_READER,
} from '../list-groups-reader';
import {
  builtInListGroups,
  listGroupsFromBriefs,
} from '../pages/lists/built-in-lists';

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
  private readonly $persistedListGroups = signal<IListGroup[]>([]);

  private readonly menuCtrl = inject(MenuController);
  private readonly router = inject(Router);
  private readonly listService = inject<IListusService>(LISTUS_SERVICE);
  private readonly listGroupsReader = inject<IListusListGroupsReader>(
    LISTUS_LIST_GROUPS_READER,
    { optional: true },
  );
  private listGroupsSubscription?: Subscription;
  /** True once an optional reader snapshot is the persisted overview. */
  private readerActive = false;
  private readerGroups?: readonly IListGroup[];

  constructor() {
    super();
    // Seed the built-in default lists (e.g. family To Buy / To Do) as soon as the
    // space type is known from the URL, before the space document loads. Mirrors
    // the lists page so the menu shows lists instantly; onSpaceDboChanged() below
    // re-seeds + merges persisted lists once the DBO arrives.
    this.spaceTypeChanged$
      .pipe(takeUntil(this.destroyed$))
      .subscribe((spaceType) => {
        if (spaceType) this.refreshListGroups();
      });
    this.spaceIDChanged$
      .pipe(
        tap(() => {
          this.readerActive = false;
          this.$persistedListGroups.set([]);
          this.refreshListGroups();
        }),
        switchMap((spaceID) =>
          spaceID ? this.listService.observeSpaceLists(spaceID) : of({}),
        ),
        takeUntil(this.destroyed$),
      )
      .subscribe({
        next: (briefs) => {
          this.$persistedListGroups.set(listGroupsFromBriefs(briefs));
          // A reader snapshot is complete, so a later extension emit must not
          // put a removed list back. Until the reader emits, show these briefs.
          if (!this.readerActive) this.refreshListGroups();
        },
        error: this.errorLogger.logErrorHandler('Failed to load lists'),
      });
  }

  // Mirror the lists page: built-in defaults (family) + the lists persisted on
  // the space, deduped by group type.
  protected override onSpaceDboChanged(): void {
    super.onSpaceDboChanged();
    this.listGroupsSubscription?.unsubscribe();
    this.listGroupsSubscription = undefined;
    if (!this.space) {
      this.readerActive = false;
      this.readerGroups = undefined;
      this.$persistedListGroups?.set([]);
      this.$listGroups.set([]);
      return;
    }
    if (!this.listGroupsReader) {
      this.refreshListGroups();
      return;
    }
    this.publishGroups(this.legacyListGroups());
    this.listGroupsSubscription = this.listGroupsReader
      .watchListGroups(this.space)
      .pipe(takeUntil(this.destroyed$))
      .subscribe({
        next: (storageGroups) => {
          this.readerActive = true;
          this.readerGroups = storageGroups;
          this.$persistedListGroups?.set([...storageGroups]);
          this.publishGroups(storageGroups);
        },
        error: (error: unknown) => {
          this.readerActive = true;
          this.readerGroups = [];
          this.$persistedListGroups?.set([]);
          this.publishGroups([]);
          this.errorLogger.logErrorHandler(
            'Failed to load storage-backed list overview',
          )(error);
        },
      });
  }

  private refreshListGroups(): void {
    if (this.readerActive) {
      this.publishGroups(this.readerGroups ?? []);
      return;
    }
    const persisted = this.$persistedListGroups?.() ?? [];
    this.publishGroups([...persisted, ...this.legacyListGroups()]);
  }

  private legacyListGroups(): readonly IListGroup[] {
    return (
      (this.space?.dbo as unknown as IListusSpaceDbo | undefined)?.listGroups ??
      []
    );
  }

  private publishGroups(persisted: readonly IListGroup[]): void {
    const groups = this.space ? [...builtInListGroups(this.space.type)] : [];
    persisted.forEach((g) => {
      const group = groups.find((current) => current.type === g.type);
      if (!group) {
        groups.push({ ...g, lists: [...(g.lists ?? [])] });
        return;
      }
      if (!group.lists) {
        group.lists = [];
      }
      group.lists.push(
        ...(g.lists ?? []).filter(
          (list) => !group.lists?.some((current) => current.id === list.id),
        ),
      );
    });
    this.$listGroups.set(groups);
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

    if (!currentList) {
      this.navigateToSelectedSpace(target, currentList);
      return;
    }

    const builtIn = builtInListGroups(target.type);
    if (this.hasListInGroups(builtIn, currentList.type, currentList.id)) {
      this.navigateToSelectedSpace(target, currentList, true);
      return;
    }

    this.listService
      .observeSpaceLists(spaceRef.id)
      .pipe(take(1), takeUntil(this.destroyed$))
      .subscribe({
        next: (briefs) =>
          this.navigateToSelectedSpace(
            target,
            currentList,
            this.hasListInGroups(
              listGroupsFromBriefs(briefs),
              currentList.type,
              currentList.id,
            ),
          ),
        error: this.errorLogger.logErrorHandler(
          'Failed to load selected space before navigating',
        ),
      });
  }

  private navigateToSelectedSpace(
    space: ISpaceContext,
    currentList?: { type: ListType; id: string },
    currentListExists = currentList
      ? this.hasListInGroups(
          this.$persistedListGroups(),
          currentList.type,
          currentList.id,
        ) ||
        this.hasListInGroups(
          builtInListGroups(space.type),
          currentList.type,
          currentList.id,
        )
      : false,
  ): void {
    const page =
      currentList && currentListExists
        ? `list/${currentList.type}/${currentList.id}`
        : 'lists';
    this.spaceNav
      .navigateForwardToSpacePage(space, page, { replaceUrl: true })
      .catch(
        this.errorLogger.logErrorHandler(
          'Failed to navigate to selected space',
        ),
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

  private hasListInGroups(
    groups: readonly IListGroup[],
    type: ListType,
    id: string,
  ): boolean {
    return groups.some((group) =>
      group.lists?.some((list) => list.type === type && list.id === id),
    );
  }
}
