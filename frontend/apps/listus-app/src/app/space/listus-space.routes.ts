import { Route } from '@angular/router';
import { listusRoutes } from '@sneat/extension-listus';
import {
  SpaceComponentBaseParams,
  SpaceMenuComponent,
} from '@sneat/space-components';

// Thin, listus-only space shell. It provides SpaceComponentBaseParams (which
// resolves the active space from the :spaceType/:spaceID route params) to all
// children, then mounts ONLY the listus routes — unlike sneat-app's
// @sneat/space-pages, which bundles every extension. This keeps listus.app
// decoupled while reusing the published @sneat/space-components context wiring.
export const listusSpaceRoutes: Route[] = [
  {
    path: '',
    providers: [SpaceComponentBaseParams],
    children: [
      {
        path: '',
        component: SpaceMenuComponent,
        outlet: 'menu',
      },
      {
        path: '',
        pathMatch: 'full',
        redirectTo: 'lists',
      },
      ...listusRoutes,
    ],
  },
];
