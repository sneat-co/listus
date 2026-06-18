import { appEnvironmentConfig } from '@sneat/app';
import { IEnvironmentConfig } from '@sneat/core';

// Single environment for listus — fail-safe by construction. appEnvironmentConfig
// returns this production config on every deployed domain and the Firebase
// emulator config only on localhost (decided at runtime from the hostname). No
// environment.prod.ts / fileReplacements: a mis-built or mis-deployed bundle can
// never point real users at the emulator.
//
// Reuses the shared sneat production Firebase project (sneat-eur3-1) — listus
// shares auth, spaces and Firestore with the rest of the sneat ecosystem.
export const listusAppEnvironmentConfig: IEnvironmentConfig =
  appEnvironmentConfig({
    production: true,
    agents: {},
    firebaseConfig: {
      projectId: 'sneat-eur3-1',
      appId: '1:588648831063:web:303af7e0c5f8a7b10d6b12',
      apiKey: 'AIzaSyCeQu1WC182yD0VHrRm4nHUxVf27fY-MLQ',
      // The Firebase Hosting site domain is same-origin with the served app and
      // is auto-authorized for OAuth (Firebase serves /__/auth/handler on
      // *.web.app), so signInWithRedirect stays first-party. Switch to
      // 'listus.app' once that custom domain + its OAuth handler are wired.
      authDomain: 'listus-app.web.app',
      messagingSenderId: '588648831063',
      measurementId: 'G-TYBDTV738R',
    },
    // Full-page redirect sign-in is the robust default for a freshly-deployed
    // domain. BaseAppComponent completes it via getRedirectResult().
    signInMethod: 'redirect',
  });
