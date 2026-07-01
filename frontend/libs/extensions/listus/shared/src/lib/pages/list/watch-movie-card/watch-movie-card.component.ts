import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  input,
  output,
  signal,
} from '@angular/core';
import {
  IonBadge,
  IonButton,
  IonCard,
  IonCardContent,
  IonIcon,
  IonSpinner,
  IonText,
  ToastController,
} from '@ionic/angular/standalone';
import {
  IListContext,
  IListItemBrief,
  ISetListItemsIsComplete,
} from '@sneat/extension-listus-contract';
import { ListusComponentBaseParams } from '../../../listus-component-base-params';
import { IListItemWithUiState } from '../list-item-with-ui-state';
import {
  formatCastLine,
  formatWatchWithLabel,
  truncateOverview,
  youTubeWatchUrl,
} from './watch-movie-card.utils';

// Renders a single watch-list item as a movie card - used only by the
// list-page "cards" segment for watch-typed lists (see
// list-page.component.html). Reuses the same IListusService calls
// (setListItemsIsCompleted / deleteListItems) that listus-list-item uses for
// the "list" segment, rather than duplicating the HTTP wiring.
@Component({
  selector: 'listus-watch-movie-card',
  templateUrl: './watch-movie-card.component.html',
  styleUrls: ['./watch-movie-card.component.scss'],
  imports: [
    IonCard,
    IonCardContent,
    IonButton,
    IonIcon,
    IonBadge,
    IonText,
    IonSpinner,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class WatchMovieCardComponent {
  private readonly params = inject(ListusComponentBaseParams);
  private readonly toastCtrl = inject(ToastController);

  public readonly $listItemWithUiState = input.required<IListItemWithUiState>();
  public readonly $list = input.required<IListContext | undefined>();

  public readonly itemChanged = output<{
    old: IListItemWithUiState;
    new: IListItemWithUiState;
  }>();

  protected readonly $isBusy = signal(false);

  protected readonly $item = computed(
    (): IListItemBrief => this.$listItemWithUiState().brief,
  );

  protected readonly $isDone = computed(() => this.$item().status === 'done');

  protected readonly $overviewSnippet = computed(() =>
    truncateOverview(this.$item().overview),
  );

  protected readonly $castLine = computed(() => formatCastLine(this.$item().cast));

  protected readonly $watchWithLabel = computed(() =>
    formatWatchWithLabel(this.$item().watchWith),
  );

  private get listService() {
    return this.params.listService;
  }

  private get errorLogger() {
    return this.params.spaceParams.errorLogger;
  }

  protected openTrailer(event: Event): void {
    event.stopPropagation();
    const key = this.$item().trailerYouTubeKey;
    if (!key) {
      return;
    }
    window.open(youTubeWatchUrl(key), '_blank', 'noopener');
  }

  protected toggleWatched(event: Event): void {
    event.stopPropagation();
    const list = this.$list();
    const item = this.$listItemWithUiState();
    if (!list?.brief || this.$isBusy()) {
      return;
    }
    const isDone = !this.$isDone();
    const newItem: IListItemWithUiState = {
      brief: { ...item.brief, status: isDone ? 'done' : undefined },
      state: { ...item.state, isChangingIsDone: true },
    };
    this.itemChanged.emit({ old: item, new: newItem });
    this.$isBusy.set(true);
    const request: ISetListItemsIsComplete = {
      spaceID: list.space.id,
      listID: list.id,
      itemIDs: [item.brief.id],
      isDone,
    };
    this.listService.setListItemsIsCompleted(request).subscribe({
      next: () => {
        this.itemChanged.emit({
          old: newItem,
          new: {
            brief: newItem.brief,
            state: { ...newItem.state, isChangingIsDone: false },
          },
        });
      },
      error: (err) => {
        this.errorLogger.logError(
          err,
          'Failed to mark movie as ' + (isDone ? 'watched' : 'to watch'),
        );
        this.showErrorToast(
          isDone
            ? 'Failed to mark movie as watched'
            : 'Failed to mark movie as to watch',
        );
        // Roll the optimistic update back.
        this.itemChanged.emit({ old: newItem, new: item });
      },
      complete: () => {
        this.$isBusy.set(false);
      },
    });
  }

  protected deleteItem(event: Event): void {
    event.stopPropagation();
    const list = this.$list();
    const item = this.$item();
    if (!list?.id || this.$isBusy()) {
      return;
    }
    if (!confirm(`Remove "${item.title}" from your watchlist?`)) {
      return;
    }
    this.$isBusy.set(true);
    this.listService
      .deleteListItems({
        spaceID: list.space.id,
        listID: list.id,
        itemIDs: [item.id],
      })
      .subscribe({
        error: (err) => {
          this.errorLogger.logError(err, 'Failed to remove movie from watchlist');
          this.showErrorToast('Failed to remove movie from watchlist');
          this.$isBusy.set(false);
        },
        complete: () => {
          this.$isBusy.set(false);
        },
      });
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
