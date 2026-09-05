import {
  ChangeDetectionStrategy,
  Component,
  inject,
  input,
  output,
  signal,
} from '@angular/core';
import { IonButton, IonIcon, IonSpinner } from '@ionic/angular';
import { addIcons } from 'ionicons';
import { refreshOutline } from 'ionicons/icons';
import {
  IApplyListTemplateRequest,
  IApplyListTemplateResult,
  IListTemplateHappeningLink,
  IListusService,
  LISTUS_SERVICE,
} from '@sneat/extension-listus-contract';

@Component({
  selector: 'listus-apply-list-template-button',
  standalone: true,
  imports: [IonButton, IonIcon, IonSpinner],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `<ion-button
      expand="block"
      [disabled]="$applying()"
      (click)="apply()"
    >
      @if ($applying()) {
        <ion-spinner name="crescent" />
      } @else {
        <ion-icon slot="start" name="refresh-outline" />
      }
      {{ label() }}</ion-button
    >
    @if ($error()) {
      <p class="ion-text-wrap ion-padding-horizontal" role="alert">
        {{ $error() }}
      </p>
    }`,
})
export class ApplyListTemplateButtonComponent {
  private readonly lists = inject<IListusService>(LISTUS_SERVICE);
  readonly spaceID = input.required<string>();
  readonly link = input.required<IListTemplateHappeningLink>();
  readonly label = input('Add regular items');
  readonly applied = output<IApplyListTemplateResult>();
  readonly $applying = signal(false);
  readonly $error = signal('');
  constructor() {
    addIcons({ refreshOutline });
  }

  apply(): void {
    if (this.$applying()) {
      return;
    }
    this.$applying.set(true);
    this.$error.set('');
    const link = this.link();
    const request: IApplyListTemplateRequest = {
      spaceID: this.spaceID(),
      sourceListID: link.sourceListID,
      destinationListID: link.destinationListID,
      requestID: crypto.randomUUID(),
    };
    this.lists.applyListTemplate(request).subscribe({
      next: (result) => {
        this.$applying.set(false);
        this.applied.emit(result);
      },
      error: (error: unknown) => {
        this.$applying.set(false);
        this.$error.set(error instanceof Error ? error.message : String(error));
      },
    });
  }
}
