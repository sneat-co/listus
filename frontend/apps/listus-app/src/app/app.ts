import { Component } from '@angular/core';
import { IonApp, IonRouterOutlet } from '@ionic/angular/standalone';
import { BaseAppComponent } from '@sneat/app';

// Extends BaseAppComponent so listus gets the shared app lifecycle: completing
// signInWithRedirect via getRedirectResult() (without this, redirect sign-in
// never finishes and the user stays unauthenticated), the document-title
// strategy, analytics pageviews, and current-space clearing.
@Component({
  selector: 'listus-root',
  template: '<ion-app><ion-router-outlet /></ion-app>',
  imports: [IonApp, IonRouterOutlet],
})
export class App extends BaseAppComponent {}
