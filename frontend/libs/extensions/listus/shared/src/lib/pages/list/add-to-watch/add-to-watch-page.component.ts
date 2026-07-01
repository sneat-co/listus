import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  signal,
} from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ParamMap } from '@angular/router';
import {
  IonBackButton,
  IonButton,
  IonButtons,
  IonCard,
  IonCardContent,
  IonCol,
  IonContent,
  IonGrid,
  IonHeader,
  IonIcon,
  IonInput,
  IonItem,
  IonItemDivider,
  IonLabel,
  IonList,
  IonRadio,
  IonRadioGroup,
  IonRow,
  IonSelect,
  IonSelectOption,
  IonSpinner,
  IonText,
  IonTitle,
  IonToolbar,
  ToastController,
} from '@ionic/angular/standalone';
import { ContactService } from '@sneat/extension-contactus-internal';
import {
  AddMovieToWatchlistRequest,
  IWatchWith,
  ListType,
  MovieSummary,
  WatchWithMode,
} from '@sneat/extension-listus-contract';
import {
  SpaceComponentBaseParams,
  SpacePageBaseComponent,
} from '@sneat/space-components';
import { ClassName } from '@sneat/ui';
import { ListusComponentBaseParams } from '../../../listus-component-base-params';

interface IContactOption {
  readonly id: string;
  readonly title: string;
}

type AddToWatchStep = 'search' | 'confirm';

// The single entry point for adding a movie to a watch-typed list. Reached
// from list-page.component.ts's newItem() ("+" header button, watch lists
// only). Flow: search (title or actor, one input) -> pick a result -> confirm
// + choose who you're watching with -> addMovieToWatchlist -> on success,
// replaceUrl back to the watch list page so the new card shows up and Back
// doesn't reopen this form. See listus-routing.ts for the route
// (list/:listType/:listID/add).
@Component({
  selector: 'listus-add-to-watch',
  templateUrl: './add-to-watch-page.component.html',
  imports: [
    FormsModule,
    IonHeader,
    IonToolbar,
    IonButtons,
    IonBackButton,
    IonTitle,
    IonContent,
    IonCard,
    IonCardContent,
    IonItem,
    IonItemDivider,
    IonLabel,
    IonInput,
    IonButton,
    IonIcon,
    IonList,
    IonRadio,
    IonRadioGroup,
    IonSelect,
    IonSelectOption,
    IonSpinner,
    IonText,
    IonGrid,
    IonRow,
    IonCol,
  ],
  providers: [
    { provide: ClassName, useValue: 'AddToWatchPageComponent' },
    SpaceComponentBaseParams,
    ListusComponentBaseParams,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class AddToWatchPageComponent extends SpacePageBaseComponent {
  private readonly params = inject(ListusComponentBaseParams);
  private readonly contactService = inject(ContactService);
  private readonly toastCtrl = inject(ToastController);

  protected readonly listType = signal<ListType | undefined>(undefined);
  protected listID?: string;

  protected readonly $step = signal<AddToWatchStep>('search');
  protected readonly $query = signal('');
  protected readonly $isSearching = signal(false);
  protected readonly $searchError = signal<string | undefined>(undefined);
  protected readonly $results = signal<MovieSummary[] | undefined>(undefined);

  protected readonly $selectedMovie = signal<MovieSummary | undefined>(
    undefined,
  );
  protected readonly $watchWithMode = signal<WatchWithMode>('alone');
  protected readonly $selectedContactID = signal<string | undefined>(
    undefined,
  );
  protected readonly $contacts = signal<IContactOption[] | undefined>(
    undefined,
  );
  protected readonly $isLoadingContacts = signal(false);
  protected readonly $isSubmitting = signal(false);

  protected readonly $canSubmit = computed(() => {
    if (this.$isSubmitting()) {
      return false;
    }
    return (
      this.$watchWithMode() !== 'contact' || !!this.$selectedContactID()
    );
  });

  private get listService() {
    return this.params.listService;
  }

  protected override onRouteParamsChanged(params: ParamMap): void {
    super.onRouteParamsChanged(params);
    const listType = params.get('listType') as ListType | null;
    const listID = params.get('listID');
    this.listType.set(listType || undefined);
    this.listID = listID || undefined;
    if (listType && listID) {
      this.$defaultBackUrlSpacePath.set(`list/${listType}/${listID}`);
    }
  }

  protected search(): void {
    const query = this.$query().trim();
    if (!query) {
      return;
    }
    this.$isSearching.set(true);
    this.$searchError.set(undefined);
    this.listService.searchMovies({ query }).subscribe({
      next: (response) => {
        this.$results.set(response.movies || []);
        this.$isSearching.set(false);
      },
      error: (err) => {
        this.errorLogger.logError(err, 'Failed to search movies');
        this.$searchError.set('Failed to search movies. Please try again.');
        this.$isSearching.set(false);
      },
    });
  }

  protected pickMovie(movie: MovieSummary): void {
    this.$selectedMovie.set(movie);
    this.$watchWithMode.set('alone');
    this.$selectedContactID.set(undefined);
    this.$step.set('confirm');
  }

  protected backToSearch(): void {
    this.$step.set('search');
    this.$selectedMovie.set(undefined);
  }

  protected onWatchWithModeChanged(mode: WatchWithMode): void {
    this.$watchWithMode.set(mode);
    if (mode === 'contact' && !this.$contacts() && !this.$isLoadingContacts()) {
      this.loadContacts();
    }
  }

  private loadContacts(): void {
    if (!this.space?.id) {
      return;
    }
    this.$isLoadingContacts.set(true);
    this.contactService.watchSpaceContacts(this.space).subscribe({
      next: (contacts) => {
        this.$contacts.set(
          contacts.map((c) => ({
            id: c.id,
            title: c.brief?.title || 'Unnamed contact',
          })),
        );
        this.$isLoadingContacts.set(false);
      },
      error: (err) => {
        this.errorLogger.logError(err, 'Failed to load contacts');
        this.$contacts.set([]);
        this.$isLoadingContacts.set(false);
      },
    });
  }

  protected submit(): void {
    const movie = this.$selectedMovie();
    if (!movie || !this.space?.id || this.$isSubmitting()) {
      return;
    }
    const watchWith = this.buildWatchWith();
    if (!watchWith) {
      return;
    }
    this.$isSubmitting.set(true);
    const request: AddMovieToWatchlistRequest = {
      spaceID: this.space.id,
      tmdbID: movie.tmdbID,
      watchWith,
    };
    this.listService.addMovieToWatchlist(request).subscribe({
      next: () => {
        this.$isSubmitting.set(false);
        const listType = this.listType() || 'watch';
        const listID = this.listID || 'movies';
        this.spaceNav
          .navigateForwardToSpacePage(
            this.space,
            `list/${listType}/${listID}`,
            { replaceUrl: true },
          )
          .catch(
            this.errorLogger.logErrorHandler(
              'Failed to navigate back to watchlist',
            ),
          );
      },
      error: (err) => {
        this.$isSubmitting.set(false);
        this.errorLogger.logError(err, 'Failed to add movie to watchlist');
        this.showErrorToast(
          'Failed to add movie to watchlist. Please try again.',
        );
      },
    });
  }

  private buildWatchWith(): IWatchWith | undefined {
    const mode = this.$watchWithMode();
    if (mode === 'alone') {
      return { mode: 'alone' };
    }
    if (mode === 'space') {
      if (!this.space?.id) {
        return undefined;
      }
      return {
        mode: 'space',
        ref: this.space.id,
        title: this.space.brief?.title,
      };
    }
    const contactID = this.$selectedContactID();
    const contact = this.$contacts()?.find((c) => c.id === contactID);
    if (!contactID || !contact) {
      this.showErrorToast('Please select who you are watching with.');
      return undefined;
    }
    return { mode: 'contact', ref: contactID, title: contact.title };
  }

  private showErrorToast(message: string): void {
    this.toastCtrl
      .create({
        message,
        duration: 2500,
        color: 'danger',
        buttons: [{ role: 'cancel', text: 'OK' }],
      })
      .then((toast) => toast.present())
      .catch(this.errorLogger.logErrorHandler('Failed to present toast'));
  }
}
