import { TestBed } from '@angular/core/testing';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { ToastController } from '@ionic/angular';
import { addIcons } from 'ionicons';
import * as ionicons from 'ionicons/icons';
import { RandomIdService } from '@sneat/random';
import {
  IListItemSourceActionNavigator,
  LIST_ITEM_SOURCE_ACTION_NAVIGATOR,
} from '@sneat/extension-listus-contract';
import {
  ISourceLinkedDateTaskReader,
  SOURCE_LINKED_DATE_TASK_READER,
} from '@sneat/extension-calendarius-contract';
import { of, throwError } from 'rxjs';
import { ListusComponentBaseParams } from '../../../listus-component-base-params';
import { ListDialogsService } from '../../dialogs/ListDialogs.service';
import { ListItemComponent } from './list-item.component';

// jsdom has no layout engine and cannot resolve the relative `svg/*.svg`
// URLs Ionicons lazily fetches for an unregistered icon name, so rendering
// any <ion-icon> used by this component's template throws `TypeError:
// Invalid URL`. Register the icons this template references up front, the
// same way the app bootstrap does in apps/listus-app/src/register-ionicons.ts.
addIcons({
  'close-outline': ionicons.closeOutline,
  'trash-outline': ionicons.trashOutline,
  checkmark: ionicons.checkmark,
  trash: ionicons.trash,
});

describe('ListItemComponent linked task authority', () => {
  const errorLogger = { logError: vi.fn(), logErrorHandler: () => vi.fn() };
  const listService = {
    setListItemsIsCompleted: vi.fn(),
    saveListItemDateTask: vi.fn(),
    deleteListItems: vi.fn(),
  };
  const sourceNavigator: IListItemSourceActionNavigator = {
    navigateToListItemSource: vi.fn(),
  };
  const dateTaskReader: ISourceLinkedDateTaskReader = {
    observeSourceLinkedDateTask: vi.fn(() =>
      of({
        happeningID: 'due-1',
        title: 'Pay electricity',
        dueDate: '2026-09-30',
        revision: 1,
        source: {
          namespace: 'listus',
          ownerSpaceID: 'space-1',
          recordID: 'do!tasks',
          lineID: 'item-1',
        },
        state: 'active' as const,
      }),
    ),
  };

  function create(item: Record<string, unknown>) {
    TestBed.configureTestingModule({
      imports: [ListItemComponent],
      providers: [
        provideNoopAnimations(),
        {
          provide: ListusComponentBaseParams,
          useValue: { listService, spaceParams: { errorLogger } },
        },
        { provide: ListDialogsService, useValue: {} },
        {
          provide: ToastController,
          useValue: { create: vi.fn().mockResolvedValue({ present: vi.fn() }) },
        },
        {
          provide: RandomIdService,
          useValue: { newRandomId: vi.fn().mockReturnValue('operation-1') },
        },
        {
          provide: LIST_ITEM_SOURCE_ACTION_NAVIGATOR,
          useValue: sourceNavigator,
        },
        { provide: SOURCE_LINKED_DATE_TASK_READER, useValue: dateTaskReader },
      ],
    });
    const fixture = TestBed.createComponent(ListItemComponent);
    fixture.componentRef.setInput('$doneFilter', 'active');
    fixture.componentRef.setInput('$listMode', 'swipe');
    fixture.componentRef.setInput('$listItemWithUiState', {
      brief: { id: 'item-1', title: 'Pay electricity', ...item },
      state: {},
    });
    fixture.componentRef.setInput('$list', {
      id: 'tasks',
      space: { id: 'space-1', type: 'family' },
      brief: { id: 'tasks', type: 'do', title: 'To do' },
    });
    fixture.detectChanges();
    return fixture.componentInstance as unknown as {
      setIsDone(value: boolean): void;
      onDueDateChanged(event: Event): void;
    };
  }

  beforeEach(() => {
    vi.clearAllMocks();
    listService.setListItemsIsCompleted.mockReturnValue(of(undefined));
    listService.saveListItemDateTask.mockReturnValue(
      of({
        itemID: 'item-1',
        dateTask: {
          happening: { module: 'calendarius', collection: 'happenings', id: 'due-1' },
          source: { module: 'listus', collection: 'lists', id: 'do!tasks' },
          purpose: 'due-date',
          revision: 2,
        },
      }),
    );
    vi.mocked(sourceNavigator.navigateToListItemSource).mockResolvedValue(true);
  });

  it('delegates source-managed completion to the owning extension', async () => {
    const component = create({
      sourceManagement: {
        source: { module: 'debtus', collection: 'sourceObligations', id: 'bill-1' },
        purpose: 'payment-due',
        completion: { disposition: 'requires_input', actionID: 'record-payment' },
      },
    });

    component.setIsDone(true);
    await Promise.resolve();

    expect(sourceNavigator.navigateToListItemSource).toHaveBeenCalledOnce();
    expect(listService.setListItemsIsCompleted).not.toHaveBeenCalled();
    expect(listService.saveListItemDateTask).not.toHaveBeenCalled();
  });

  it('uses the Calendar-coordinated endpoint and keeps the retry operation ID', () => {
    const component = create({
      dateTask: {
        happening: { module: 'calendarius', collection: 'happenings', id: 'due-1' },
        source: { module: 'listus', collection: 'lists', id: 'do!tasks' },
        purpose: 'due-date',
        revision: 1,
      },
    });
    listService.saveListItemDateTask.mockReturnValueOnce(
      throwError(() => new Error('network')),
    );

    component.setIsDone(true);
    component.setIsDone(true);

    const [first, second] = listService.saveListItemDateTask.mock.calls;
    expect(first[0].operationID).toBe('operation-1');
    expect(second[0]).toEqual(first[0]);
    expect(listService.setListItemsIsCompleted).not.toHaveBeenCalled();
  });

  it('adds a due date through the Calendar-coordinated endpoint', () => {
    const component = create({});
    const input = document.createElement('input');
    input.value = '2026-09-30';

    component.onDueDateChanged({
      target: input,
      preventDefault: vi.fn(),
      stopPropagation: vi.fn(),
    } as unknown as Event);

    expect(listService.saveListItemDateTask).toHaveBeenCalledWith({
      spaceID: 'space-1',
      listID: 'do!tasks',
      itemID: 'item-1',
      operationID: 'operation-1',
      expectedTaskRevision: 0,
      dueDate: '2026-09-30',
      state: 'active',
    });
    expect(listService.setListItemsIsCompleted).not.toHaveBeenCalled();
  });

  it('reopens with the Calendar-owned due date loaded after a cold read', () => {
    const component = create({
      status: 'done',
      dateTask: {
        happening: { module: 'calendarius', collection: 'happenings', itemID: 'due-1' },
        source: { module: 'listus', collection: 'lists', itemID: 'do!tasks' },
        purpose: 'due-date',
        revision: 1,
      },
    });

    component.setIsDone(false);

    expect(listService.saveListItemDateTask).toHaveBeenCalledWith(
      expect.objectContaining({
        expectedTaskRevision: 1,
        dueDate: '2026-09-30',
        state: 'active',
      }),
    );
    expect(listService.setListItemsIsCompleted).not.toHaveBeenCalled();
  });
});
