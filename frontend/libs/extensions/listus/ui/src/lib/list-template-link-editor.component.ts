import {
  ChangeDetectionStrategy,
  Component,
  computed,
  input,
  output,
} from '@angular/core';
import { FormsModule } from '@angular/forms';
import { IonButton } from '@ionic/angular/ion-button';
import { IonItem } from '@ionic/angular/ion-item';
import { IonList } from '@ionic/angular/ion-list';
import { IonSelect } from '@ionic/angular/ion-select';
import { IonSelectOption } from '@ionic/angular/ion-select-option';
import {
  IListInfo,
  IListTemplateHappeningLink,
  ListType,
} from '@sneat/extension-listus-contract';

@Component({
  selector: 'listus-list-template-link-editor',
  standalone: true,
  imports: [
    FormsModule,
    IonButton,
    IonItem,
    IonList,
    IonSelect,
    IonSelectOption,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <ion-list lines="full">
      <ion-item>
        <ion-select
          label="Regular items"
          labelPlacement="stacked"
          placeholder="Choose a saved list"
          interface="action-sheet"
          [ngModel]="value()?.sourceListID"
          (ngModelChange)="selectSource($event)"
        >
          @for (list of $templates(); track list.id) {
            <ion-select-option [value]="list.id"
              >{{ list.emoji }} {{ list.title }}</ion-select-option
            >
          }
        </ion-select>
      </ion-item>
      <ion-item>
        <ion-select
          label="Add to"
          labelPlacement="stacked"
          placeholder="Choose the ongoing list"
          interface="action-sheet"
          [disabled]="!value()?.sourceListID"
          [ngModel]="value()?.destinationListID"
          (ngModelChange)="selectDestination($event)"
        >
          @for (list of $destinations(); track list.id) {
            <ion-select-option [value]="list.id"
              >{{ list.emoji }} {{ list.title }}</ion-select-option
            >
          }
        </ion-select>
      </ion-item>
    </ion-list>
    <ion-button
      fill="clear"
      size="small"
      (click)="createTemplate.emit(listType())"
    >
      Create a regular-items list
    </ion-button>
    <p class="ion-padding-horizontal ion-no-margin">
      Maintain regular items in an ordinary Listus list. Applying it adds
      missing items and restores completed ones.
    </p>
  `,
})
export class ListTemplateLinkEditorComponent {
  readonly lists = input.required<readonly IListInfo[]>();
  readonly listType = input.required<Extract<ListType, 'buy' | 'do'>>();
  readonly value = input<IListTemplateHappeningLink>();
  readonly valueChange = output<IListTemplateHappeningLink | undefined>();
  readonly createTemplate = output<Extract<ListType, 'buy' | 'do'>>();

  readonly $templates = computed(() => this.eligibleLists());
  readonly $destinations = computed(() => {
    const sourceID = this.value()?.sourceListID;
    return this.eligibleLists().filter((list) => list.id !== sourceID);
  });

  selectSource(sourceListID: string): void {
    const current = this.value();
    const destinationStillValid =
      current?.destinationListID !== sourceListID &&
      this.eligibleLists().some(
        (list) => list.id === current?.destinationListID,
      );
    this.valueChange.emit({
      sourceListID,
      destinationListID: destinationStillValid
        ? current!.destinationListID
        : '',
    });
  }

  selectDestination(destinationListID: string): void {
    const sourceListID = this.value()?.sourceListID;
    if (sourceListID) {
      this.valueChange.emit({ sourceListID, destinationListID });
    }
  }

  private eligibleLists(): readonly IListInfo[] {
    return this.lists().filter(
      (list) => list.type === this.listType() && !!list.id && !list.hidden,
    );
  }
}
