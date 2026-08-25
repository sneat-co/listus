import {
  ChangeDetectionStrategy,
  Component,
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
  IonContent,
  IonHeader,
  IonIcon,
  IonItem,
  IonLabel,
  IonList,
  IonSpinner,
  IonText,
  IonTextarea,
  IonTitle,
  IonToolbar,
  ToastController,
} from '@ionic/angular';
import {
  AddMovieToWatchlistRequest,
  ListType,
  MovieSummary,
} from '@sneat/extension-listus-contract';
import {
  SpaceComponentBaseParams,
  SpacePageBaseComponent,
} from '@sneat/space-components';
import { ClassName } from '@sneat/ui';
import { ListusComponentBaseParams } from '../../../listus-component-base-params';

// Discover page for watch lists: describe a movie in your own words ("that
// submarine movie with the guy from Titanic") and get TMDB candidates via the
// AI-grounded `listus/movies/identify` endpoint, then add a candidate straight
// to the watch list this page was opened from. Reached from the list page's
// footer "Discover" button / discover segment (route
// `list/:listType/:listID/discover` - see listus-routing.ts).
@Component({
  selector: 'listus-discover',
  templateUrl: './discover-page.component.html',
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
    IonLabel,
    IonTextarea,
    IonButton,
    IonIcon,
    IonList,
    IonSpinner,
    IonText,
  ],
  providers: [
    { provide: ClassName, useValue: 'DiscoverPageComponent' },
    SpaceComponentBaseParams,
    ListusComponentBaseParams,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DiscoverPageComponent extends SpacePageBaseComponent {
  private readonly params = inject(ListusComponentBaseParams);
  private readonly toastCtrl = inject(ToastController);

  protected readonly $listType = signal<ListType | undefined>(undefined);
  protected listID?: string;

  protected readonly $description = signal('');
  protected readonly $isIdentifying = signal(false);
  protected readonly $identifyError = signal<string | undefined>(undefined);
  protected readonly $results = signal<MovieSummary[] | undefined>(undefined);

  // tmdbID of the candidate currently being added (disables its button) and
  // the set of candidates already added (turns the button into a checkmark).
  protected readonly $addingTmdbID = signal<number | undefined>(undefined);
  protected readonly $addedTmdbIDs = signal<ReadonlySet<number>>(new Set());

  private get listService() {
    return this.params.listService;
  }

  protected override onRouteParamsChanged(params: ParamMap): void {
    super.onRouteParamsChanged(params);
    const listType = params.get('listType') as ListType | null;
    const listID = params.get('listID');
    this.$listType.set(listType || undefined);
    this.listID = listID || undefined;
    if (listType && listID) {
      this.$defaultBackUrlSpacePath.set(`list/${listType}/${listID}`);
    }
  }

  protected identify(): void {
    const description = this.$description().trim();
    if (!description || this.$isIdentifying()) {
      return;
    }
    this.$isIdentifying.set(true);
    this.$identifyError.set(undefined);
    this.listService.identifyMovies({ description }).subscribe({
      next: (response) => {
        this.$results.set(response.movies || []);
        this.$isIdentifying.set(false);
      },
      error: (err) => {
        this.errorLogger.logError(err, 'Failed to identify movies');
        this.$identifyError.set(
          'Failed to identify movies. Please try again.',
        );
        this.$isIdentifying.set(false);
      },
    });
  }

  protected addToWatchlist(movie: MovieSummary): void {
    if (!this.space?.id || this.$addingTmdbID() !== undefined) {
      return;
    }
    const listType = this.$listType();
    const listID = this.listID;
    this.$addingTmdbID.set(movie.tmdbID);
    const request: AddMovieToWatchlistRequest = {
      spaceID: this.space.id,
      // Target the watch list this page was opened from - omitting listID
      // would default to the canonical watch!movies list server-side.
      listID: listType && listID ? `${listType}!${listID}` : undefined,
      tmdbID: movie.tmdbID,
    };
    this.listService.addMovieToWatchlist(request).subscribe({
      next: () => {
        this.$addingTmdbID.set(undefined);
        this.$addedTmdbIDs.set(
          new Set([...this.$addedTmdbIDs(), movie.tmdbID]),
        );
        this.showToast(`Added "${movie.title}" to the watchlist.`, 'success');
      },
      error: (err) => {
        this.$addingTmdbID.set(undefined);
        this.errorLogger.logError(err, 'Failed to add movie to watchlist');
        this.showToast(
          'Failed to add movie to watchlist. Please try again.',
          'danger',
        );
      },
    });
  }

  private showToast(message: string, color: 'success' | 'danger'): void {
    this.toastCtrl
      .create({
        message,
        duration: color === 'success' ? 2000 : 3500,
        color,
        buttons: [{ role: 'cancel', text: 'OK' }],
      })
      .then((toast) => toast.present())
      .catch(this.errorLogger.logErrorHandler('Failed to present toast'));
  }
}
