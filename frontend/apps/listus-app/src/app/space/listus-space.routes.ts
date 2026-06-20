import { Route } from '@angular/router';
import {
  listusRoutes,
  ListusSpaceMenuComponent,
} from '@sneat/extension-listus-shared';
import { SpaceComponentBaseParams } from '@sneat/space-components';

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
        // listus-specific side menu (space selector + the space's lists) instead
        // of the generic SpaceMenuComponent, which hardcodes every sneat-app
        // extension (Assets, Budget, Contacts, …) — none of which exist here.
        path: '',
        component: ListusSpaceMenuComponent,
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
