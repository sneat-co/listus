// Main entry point for listus.app
import { bootstrapApplication } from '@angular/platform-browser';
import { provideRouter } from '@angular/router';
import {
  getStandardSneatProviders,
  provideAppInfo,
  provideRolesByType,
} from '@sneat/app';
import { authRoutes } from '@sneat/auth-ui';
import { provideListusInternal } from '@sneat/extension-listus-internal';
import { App } from './app/app';
import { appRoutes } from './app/app.routes';
import { listusAppEnvironmentConfig } from './environments/environment';
import { registerIonicons } from './register-ionicons';

bootstrapApplication(App, {
  providers: [
    ...getStandardSneatProviders(listusAppEnvironmentConfig),
    // Bind the listus contract token (LISTUS_SERVICE) to its concrete
    // implementation. The app is the composition root and may wire -internal.
    ...provideListusInternal(),
    provideAppInfo({ appId: 'listus', appTitle: 'Listus.app' }),
    provideRouter([...appRoutes, ...authRoutes]),
    provideRolesByType(undefined),
  ],
}).catch((err) => console.error(err));

registerIonicons();
