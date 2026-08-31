// Main entry point for listus.app
import { bootstrapApplication } from '@angular/platform-browser';
import { provideRouter } from '@angular/router';
import { SneatApiBaseUrl } from '@sneat/api';
import {
  getStandardSneatProviders,
  provideAppInfo,
  provideRolesByType,
} from '@sneat/app';
import { authRoutes } from '@sneat/auth-ui';
import { provideListus } from '@sneat/extension-listus';
import { App } from './app/app';
import { appRoutes } from './app/app.routes';
import { listusAppEnvironmentConfig } from './environments/environment';
import { getListusApiBaseUrl } from './environments/listus-api-base-url';
import { registerIonicons } from './register-ionicons';

const listusApiBaseUrl = getListusApiBaseUrl(
  !!listusAppEnvironmentConfig.firebaseConfig.emulator,
);

bootstrapApplication(App, {
  providers: [
    ...getStandardSneatProviders(listusAppEnvironmentConfig),
    ...(listusApiBaseUrl
      ? [{ provide: SneatApiBaseUrl, useValue: listusApiBaseUrl }]
      : []),
    // Bind the listus contract token (LISTUS_SERVICE) to its concrete
    // implementation. The app is the composition root and may wire -internal.
    ...provideListus(),
    provideAppInfo({ appId: 'listus', appTitle: 'Listus.app' }),
    provideRouter([...appRoutes, ...authRoutes]),
    provideRolesByType(undefined),
  ],
}).catch((err) => console.error(err));

registerIonicons();
