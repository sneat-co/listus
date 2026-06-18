import { TitleCasePipe } from '@angular/common';
import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  signal,
} from '@angular/core';
import { RouterLink } from '@angular/router';
import {
  IonIcon,
  IonItem,
  IonLabel,
  IonList,
  IonNote,
  IonSelect,
  IonSelectOption,
  MenuController,
} from '@ionic/angular/standalone';
import { ISneatUserState } from '@sneat/auth-core';
import { IUserSpaceBrief } from '@sneat/auth-models';
import { AuthMenuItemComponent } from '@sneat/auth-ui';
import { ContactusServicesModule } from '@sneat/contactus-services';
import { IIdAndBrief } from '@sneat/core';
import {
  SpaceBaseComponent,
  SpaceComponentBaseParams,
} from '@sneat/space-components';
import { SpaceServiceModule } from '@sneat/space-services';
import { zipMapBriefsWithIDs } from '@sneat/space-models';
import { ClassName } from '@sneat/ui';
import { takeUntil } from 'rxjs/operators';
import { IListGroup, IListusSpaceDbo } from '../dto';
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
    TitleCasePipe,
    RouterLink,
    ContactusServicesModule,
    SpaceServiceModule,
    IonList,
    IonItem,
    IonSelect,
    IonSelectOption,
    IonIcon,
    IonLabel,
    IonNote,
    AuthMenuItemComponent,
  ],
  providers: [
    { provide: ClassName, useValue: 'ListusSpaceMenuComponent' },
    SpaceComponentBaseParams,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ListusSpaceMenuComponent extends SpaceBaseComponent {
  protected readonly $spaces = signal<
    readonly IIdAndBrief<IUserSpaceBrief>[] | undefined
  >(undefined);
  protected readonly $disabled = computed(() => !this.$spaceID());
  protected readonly $listGroups = signal<IListGroup[]>([]);

  private readonly menuCtrl = inject(MenuController);

  constructor() {
    super();
    this.spaceParams.userService.userState
      .pipe(takeUntil(this.destroyed$))
      .subscribe({
        next: (userState: ISneatUserState) =>
          this.$spaces.set(
            userState?.record
              ? zipMapBriefsWithIDs(userState.record.spaces) || []
              : undefined,
          ),
        error: this.errorLogger.logErrorHandler('failed to get user state'),
      });
  }

  // Mirror the lists page: built-in defaults (family) + the lists persisted on
  // the space DBO, deduped by group type.
  protected override onSpaceDboChanged(): void {
    super.onSpaceDboChanged();
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
  }

  protected onSpaceSelected(event: Event): void {
    const spaceID = (event as CustomEvent).detail.value as string;
    if (spaceID === this.space?.id) {
      return;
    }
    const space = this.$spaces()?.find((t) => t.id === spaceID);
    if (space) {
      this.setSpaceRef(space);
      this.spaceNav
        .navigateToSpace(space)
        .catch(
          this.errorLogger.logErrorHandler(
            'Failed to navigate to selected space',
          ),
        );
    }
    this.closeMenu();
  }

  protected closeMenu(): void {
    this.menuCtrl.close().catch(this.errorLogger.logError);
  }
}
