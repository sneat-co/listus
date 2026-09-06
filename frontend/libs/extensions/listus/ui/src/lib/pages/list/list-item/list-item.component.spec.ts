import { TestBed } from '@angular/core/testing';
import { ToastController } from '@ionic/angular';
import { RandomIdService } from '@sneat/random';
import {
  IListItemSourceActionNavigator,
  LIST_ITEM_SOURCE_ACTION_NAVIGATOR,
} from '@sneat/extension-listus-contract';
import { of, throwError } from 'rxjs';
import { ListusComponentBaseParams } from '../../../listus-component-base-params';
import { ListDialogsService } from '../../dialogs/ListDialogs.service';
import { ListItemComponent } from './list-item.component';

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

  function create(item: Record<string, unknown>) {
    TestBed.configureTestingModule({
      imports: [ListItemComponent],
      providers: [
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
      id: 'do!tasks',
      space: { id: 'space-1', type: 'family' },
      brief: { id: 'do!tasks', type: 'do', title: 'To do' },
    });
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
});
