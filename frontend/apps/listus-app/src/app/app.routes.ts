import { Route } from '@angular/router';

export const appRoutes: Route[] = [
  {
    path: '',
    pathMatch: 'full',
    redirectTo: 'login',
  },
  {
    // Space-scoped routes host the listus pages, mirroring sneat-app's
    // space/:spaceType/:spaceID mount point.
    path: 'space/:spaceType/:spaceID',
    loadChildren: () =>
      import('./space/listus-space.routes').then((m) => m.listusSpaceRoutes),
  },
];
