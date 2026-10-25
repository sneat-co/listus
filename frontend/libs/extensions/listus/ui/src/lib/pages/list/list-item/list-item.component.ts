import {
  ChangeDetectionStrategy,
  Component,
  computed,
  effect,
  input,
  output,
  signal,
  inject,
} from '@angular/core';
import {
  ToastController,
  IonBadge,
  IonButton,
  IonButtons,
  IonCheckbox,
  IonIcon,
  IonItem,
  IonItemOption,
  IonItemOptions,
  IonItemSliding,
  IonLabel,
  IonReorder,
  IonSpinner,
  IonText,
} from '@ionic/angular';
import type { ToastOptions } from '@ionic/core';
import { listItemAnimations } from '@sneat/core';
import {
  IListContext,
  IListItemBrief,
  IListItemIDsRequest,
  IListusService,
  ISetListItemsIsComplete,
  ISaveListItemDateTaskRequest,
  IListItemSourceActionNavigator,
  LIST_ITEM_SOURCE_ACTION_NAVIGATOR,
} from '@sneat/extension-listus-contract';
import {
  ISourceLinkedDateTask,
  ISourceLinkedDateTaskReader,
  SOURCE_LINKED_DATE_TASK_READER,
} from '@sneat/extension-calendarius-contract';
import { IMediaTarget } from '@sneat/extension-media-contract';
import { MediaEditorComponent, MediaImageComponent } from '@sneat/media';
import { RandomIdService } from '@sneat/random';
import { ListusComponentBaseParams } from '../../../listus-component-base-params';
import { ListDialogsService } from '../../dialogs/ListDialogs.service';
import { IListItemWithUiState } from '../list-item-with-ui-state';

@Component({
  selector: 'listus-list-item',
  imports: [
    IonItemSliding,
    IonItem,
    IonCheckbox,
    IonLabel,
    IonText,
    IonBadge,
    IonButtons,
    IonButton,
    IonIcon,
    IonSpinner,
    IonReorder,
    IonItemOptions,
    IonItemOption,
    MediaEditorComponent,
    MediaImageComponent,
  ],
  templateUrl: './list-item.component.html',
  styleUrls: ['./list-item.component.scss'],
  animations: [listItemAnimations],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ListItemComponent {
  private readonly params = inject(ListusComponentBaseParams);
  private readonly listDialogs = inject(ListDialogsService);
  private readonly toastCtrl = inject(ToastController);
  private readonly randomID = inject(RandomIdService);
  private readonly sourceActionNavigator =
    inject<IListItemSourceActionNavigator>(LIST_ITEM_SOURCE_ACTION_NAVIGATOR, {
      optional: true,
    });
  private readonly dateTaskReader = inject<ISourceLinkedDateTaskReader>(
    SOURCE_LINKED_DATE_TASK_READER,
    { optional: true },
  );

  public readonly showDoneCheckbox = input(false);

  public readonly $doneFilter = input.required<
    'all' | 'active' | 'completed'
  >();

  public readonly $listMode = input.required<'reorder' | 'swipe'>();
  protected readonly $isReorderMode = computed(
    () => this.$listMode() === 'reorder',
  );

  readonly $listItemWithUiState = input.required<IListItemWithUiState>();
  public readonly $list = input.required<IListContext | undefined>();
  protected readonly $supportsDateTask = computed(
    () => this.$list()?.brief?.type === 'do',
  );

  protected readonly $isSettingIsDone = signal(false);
  protected readonly $photoPresentation = signal<
    'thumbnail' | 'preview' | 'large'
  >('thumbnail');
  protected readonly $dateTask = signal<ISourceLinkedDateTask | undefined>(
    undefined,
  );
  protected readonly $dateTaskUnavailable = signal(false);
  private failedDateTaskRequest?: ISaveListItemDateTaskRequest;

  public readonly itemClicked = output<IListItemBrief>();

  public readonly itemChanged = output<{
    old: IListItemWithUiState;
    new: IListItemWithUiState;
  }>();

  public readonly listChanged = output<IListContext>();

  protected readonly $listItem = computed(
    () => this.$listItemWithUiState().brief,
  );
  protected readonly $photoTarget = computed<IMediaTarget | undefined>(() => {
    const list = this.$list();
    const item = this.$listItem();
    if (!list || !item.id) {
      return undefined;
    }
    return {
      scope: 'space',
      spaceID: list.space.id,
      type: 'list_item',
      id: item.id,
      parentID: canonicalListID(list),
    };
  });

  private readonly observeDateTask = effect((onCleanup) => {
    const list = this.$list();
    const item = this.$listItemWithUiState().brief;
    const listID = list ? canonicalListID(list) : undefined;
    const happeningID = item.dateTask?.happening.itemID;
    this.$dateTask.set(undefined);
    this.$dateTaskUnavailable.set(false);
    if (!list || !happeningID) {
      return;
    }
    if (!this.dateTaskReader) {
      this.$dateTaskUnavailable.set(true);
      return;
    }
    const subscription = this.dateTaskReader
      .observeSourceLinkedDateTask(list.space.id, happeningID)
      .subscribe({
        next: (task) => {
          const isExpectedTask =
            task?.happeningID === happeningID &&
            task.source.namespace === 'listus' &&
            task.source.ownerSpaceID === list.space.id &&
            task.source.recordID === listID &&
            task.source.lineID === item.id;
          this.$dateTask.set(isExpectedTask ? task : undefined);
          this.$dateTaskUnavailable.set(!isExpectedTask);
        },
        error: (err) => {
          this.$dateTask.set(undefined);
          this.$dateTaskUnavailable.set(true);
          this.errorLogger.logError(err, 'Failed to load the linked due task');
        },
      });
    onCleanup(() => subscription.unsubscribe());
  });

  private get listService(): IListusService {
    return this.params.listService;
  }

  private get errorLogger() {
    return this.params.spaceParams.errorLogger;
  }

  protected isSpinning(): boolean {
    if (!this.$listItemWithUiState) {
      return false;
    }
    const { state } = this.$listItemWithUiState();
    return (
      !!state.isReordering || !!state.isDeleting || !!state.isChangingIsDone
    );
  }

  protected goListItem(): void {
    const listItem = this.$listItem();
    console.log(
      `goListItem(${listItem?.id}), subListId=${listItem?.subListId}`,
    );
    this.itemClicked.emit(listItem);
  }

  protected $isDone = computed(
    () => !!this.$listItemWithUiState().brief.isDone,
  );

  protected isDone(item?: IListItemWithUiState): boolean {
    return !!item?.brief.isDone;
  }

  protected onIsDoneCheckboxChanged(event: Event): void {
    event.stopPropagation();
    event.preventDefault();
    if (!this.$listItemWithUiState) {
      return;
    }
    const { checked } = (event as CustomEvent).detail;
    if (checked === undefined) {
      return;
    }
    const isDone = !!checked;
    this.setIsDone(isDone);
  }

  protected onPhotoChanged(mediaID: string | undefined): void {
    const old = this.$listItemWithUiState();
    this.itemChanged.emit({
      old,
      new: {
        brief: {
          ...old.brief,
          photo: mediaID ? { mediaID } : undefined,
        },
        state: old.state,
      },
    });
    if (!mediaID) {
      this.$photoPresentation.set('thumbnail');
    }
  }

  protected togglePhotoPreview(event: Event): void {
    event.preventDefault();
    event.stopPropagation();
    this.$photoPresentation.update((presentation) => {
      switch (presentation) {
        case 'thumbnail':
          return 'preview';
        case 'preview':
          return 'large';
        case 'large':
          return 'thumbnail';
      }
    });
  }

  protected setIsDone(isDone?: boolean, ionSliding?: IonItemSliding): void {
    const item = this.$listItemWithUiState();
    if (isDone === undefined) {
      isDone = !this.$isDone();
    }
    if (item.brief.sourceManagement) {
      this.openSourceAction();
      return;
    }
    if (item.brief.dateTask) {
      if (!isDone) {
        const dueDate = this.$dateTask()?.dueDate;
        if (!dueDate) {
          this.showError('The due date could not be loaded. Try again.');
          return;
        }
        this.saveDateTask(item, 'active', dueDate);
      } else {
        this.saveDateTask(item, 'completed');
      }
      return;
    }
    const newItem: IListItemWithUiState = {
      brief: { ...item.brief, status: isDone ? 'done' : undefined },
      state: { ...item.state, isChangingIsDone: true },
    };
    const performSetIsDone = (): void => {
      this.itemChanged.emit({
        old: item,
        new: newItem,
      });

      this.$isSettingIsDone.set(true);

      const list = this.$list();
      if (!list?.brief) {
        return;
      }

      const request: ISetListItemsIsComplete = {
        spaceID: list.space.id,
        listID: canonicalListID(list),
        itemIDs: [item.brief.id],
        isDone: isDone,
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
          const toastOptions: ToastOptions = {
            message: isDone
              ? `${item.brief.title} marked as completed`
              : `${item.brief.title} marked as active`,
            duration: 1000,
            color: 'light',
            buttons: [{ icon: 'close', role: 'cancel' }],
            keyboardClose: true,
          };
          this.toastCtrl
            .create(toastOptions)
            .then((toast) =>
              toast
                .present()
                .catch(
                  this.errorLogger.logErrorHandler(
                    'Failed to present a toast message about list item isCompleted set to ' +
                      isDone,
                  ),
                ),
            )
            .catch(
              this.errorLogger.logErrorHandler(
                'Failed to present a toast message about list item isCompleted set to ' +
                  isDone,
              ),
            );
        },
        error: (err) => {
          // Roll back the optimistic state so the item stays usable when the
          // save fails instead of being left disabled in the active list.
          this.itemChanged.emit({ old: newItem, new: item });
          this.$isSettingIsDone.set(false);
          this.errorLogger.logError(
            err,
            'failed to mark list item as completed',
          );
        },
        complete: () => {
          this.$isSettingIsDone.set(false);
        },
      });
    };
    if (ionSliding) {
      (ionSliding as IonItemSliding)
        .close()
        .then(performSetIsDone)
        .catch(this.errorLogger.logErrorHandler('Failed to set completed'));
    } else {
      performSetIsDone();
      // setTimeout(() => performSetIsDone(), 0);
    }
  }

  protected openSourceAction(event?: Event): void {
    event?.preventDefault();
    event?.stopPropagation();
    const list = this.$list();
    const item = this.$listItem();
    if (!list || !this.sourceActionNavigator) {
      this.showError('This item must be updated from its source.');
      return;
    }
    this.$isSettingIsDone.set(true);
    this.sourceActionNavigator
      .navigateToListItemSource(list.space, item)
      .then((handled) => {
        if (!handled) {
          this.showError('The source for this item is unavailable.');
        }
      })
      .catch((err) => this.errorLogger.logError(err, 'Failed to open source'))
      .finally(() => this.$isSettingIsDone.set(false));
  }

  protected onDueDateChanged(event: Event): void {
    event.preventDefault();
    event.stopPropagation();
    const dueDate = (event.target as HTMLInputElement).value;
    if (!dueDate) {
      return;
    }
    this.saveDateTask(this.$listItemWithUiState(), 'active', dueDate);
  }

  protected removeDueDate(event: Event): void {
    event.preventDefault();
    event.stopPropagation();
    this.saveDateTask(this.$listItemWithUiState(), 'canceled');
  }

  private saveDateTask(
    item: IListItemWithUiState,
    state: 'active' | 'completed' | 'canceled',
    dueDate?: string,
  ): void {
    const list = this.$list();
    const dateTask = item.brief.dateTask;
    if (!list) {
      return;
    }
    const revision = dateTask?.revision ?? 0;
    const currentKey = `${list.space.id}|${list.id}|${item.brief.id}|${revision}|${state}|${dueDate ?? ''}`;
    let request = this.failedDateTaskRequest;
    if (!request || this.dateTaskRequestKey(request) !== currentKey) {
      request = {
        spaceID: list.space.id,
        listID: canonicalListID(list),
        itemID: item.brief.id,
        operationID: this.randomID.newRandomId({ len: 20 }),
        expectedTaskRevision: revision,
        dueDate,
        state,
      };
    }
    this.$isSettingIsDone.set(true);
    this.failedDateTaskRequest = request;
    this.listService.saveListItemDateTask(request).subscribe({
      next: (response) => {
        this.failedDateTaskRequest = undefined;
        this.itemChanged.emit({
          old: item,
          new: {
            brief: {
              ...item.brief,
              status:
                state === 'completed'
                  ? 'done'
                  : state === 'active'
                    ? 'active'
                    : item.brief.status,
              dateTask: response.dateTask,
            },
            state: { ...item.state, isChangingIsDone: false },
          },
        });
      },
      error: (err) => {
        this.errorLogger.logError(err, 'Failed to update the linked due task');
        this.$isSettingIsDone.set(false);
      },
      complete: () => this.$isSettingIsDone.set(false),
    });
  }

  private dateTaskRequestKey(request: ISaveListItemDateTaskRequest): string {
    return `${request.spaceID}|${request.listID}|${request.itemID}|${request.expectedTaskRevision}|${request.state}|${request.dueDate ?? ''}`;
  }

  private showError(message: string): void {
    this.toastCtrl
      .create({ message, duration: 2500, color: 'warning' })
      .then((toast) => toast.present())
      .catch(this.errorLogger.logError);
  }

  protected deleteFromList(
    item: IListItemBrief,
    ionSliding?: IonItemSliding | HTMLElement,
  ): void {
    if (!item.id) {
      return;
    }
    const list = this.$list();
    if (!list?.id || !list?.brief) {
      return;
    }
    const request: IListItemIDsRequest = {
      spaceID: list.space.id,
      listID: list.id,
      // listType: this.list?.brief?.type,
      itemIDs: [item.id],
    };
    this.listService.deleteListItems(request).subscribe({
      next: () => {
        // this.listChanged.emit(listDto);
      },
      error: this.errorLogger.logError,
      complete: () => {
        if (ionSliding) {
          (ionSliding as IonItemSliding)
            ?.closeOpened()
            .catch(this.errorLogger.logError);
        }
      },
    });
  }

  protected confirmDeleteFromList(
    item: IListItemBrief,
    event: Event,
    ionSliding?: IonItemSliding | HTMLElement,
  ): void {
    event.preventDefault();
    event.stopPropagation();
    if (item.sourceManagement) {
      this.openSourceAction(event);
      return;
    }
    if (item.dateTask) {
      this.showError('Remove the due date before deleting this item.');
      return;
    }
    if (!confirm(`Remove "${item.title}" from this list?`)) {
      return;
    }
    this.deleteFromList(item, ionSliding);
  }

  protected openCopyListItemDialog(
    listItem: IListItemBrief,
    event: Event,
  ): void {
    event.stopPropagation();

    this.listDialogs
      .copyListItems([listItem], {
        type: listItem.subListType || 'other',
        id: listItem.subListId,
        title: listItem.title,
      })
      .catch(this.errorLogger.logError);
  }
}

function canonicalListID(list: IListContext): string {
  return list.id.includes('!') ? list.id : `${list.brief?.type}!${list.id}`;
}
