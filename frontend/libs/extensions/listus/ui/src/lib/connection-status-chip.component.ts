import { DOCUMENT, isPlatformBrowser } from '@angular/common';
import {
  ChangeDetectionStrategy,
  Component,
  DestroyRef,
  PLATFORM_ID,
  inject,
  signal,
} from '@angular/core';
import { IonChip, IonLabel } from '@ionic/angular';

@Component({
  selector: 'listus-connection-status-chip',
  imports: [IonChip, IonLabel],
  template: `
    <ion-chip
      [class.offline]="!isOnline()"
      [class.online]="isOnline()"
      [attr.aria-label]="isOnline() ? 'Online' : 'Offline'"
    >
      <ion-label>{{ isOnline() ? 'Online' : 'Offline' }}</ion-label>
    </ion-chip>
  `,
  styles: [
    `
      ion-chip.online {
        --background: transparent;
        --color: var(--ion-color-success);
      }

      ion-chip.offline {
        --background: var(--ion-color-warning);
        --color: var(--ion-color-warning-contrast);
      }
    `,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ConnectionStatusChipComponent {
  private readonly destroyRef = inject(DestroyRef);
  private readonly platformId = inject(PLATFORM_ID);
  private readonly document = inject(DOCUMENT);

  protected readonly isOnline = signal(true);

  constructor() {
    const view = this.document.defaultView;
    if (!view || !isPlatformBrowser(this.platformId)) {
      return;
    }

    const updateConnectionStatus = (): void => {
      this.isOnline.set(view.navigator.onLine);
    };

    updateConnectionStatus();
    view.addEventListener('online', updateConnectionStatus);
    view.addEventListener('offline', updateConnectionStatus);
    this.destroyRef.onDestroy(() => {
      view.removeEventListener('online', updateConnectionStatus);
      view.removeEventListener('offline', updateConnectionStatus);
    });
  }
}
