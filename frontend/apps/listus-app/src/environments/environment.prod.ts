import { IEnvironmentConfig, IFirebaseConfig } from '@sneat/core';

// Production listus config. Reuses the shared sneat production Firebase project
// (sneat-eur3-1) — listus shares auth, spaces and Firestore with the rest of the
// sneat ecosystem. Swapped in for environment.ts at build time via the
// production fileReplacements in project.json.
const firebaseConfig: IFirebaseConfig = {
  projectId: 'sneat-eur3-1',
  appId: '1:588648831063:web:303af7e0c5f8a7b10d6b12',
  apiKey: 'AIzaSyCeQu1WC182yD0VHrRm4nHUxVf27fY-MLQ',
  // The Firebase Hosting site domain is same-origin with the served app and is
  // auto-authorized for OAuth (Firebase serves /__/auth/handler on *.web.app),
  // so the signInWithRedirect flow stays first-party here. When the listus.app
  // custom domain is connected, switch this to 'listus.app' (and register
  // https://listus.app/__/auth/handler on the OAuth client).
  authDomain: 'listus-app.web.app',
  messagingSenderId: '588648831063',
  measurementId: 'G-TYBDTV738R',
};

export const listusAppEnvironmentConfig: IEnvironmentConfig = {
  production: true,
  agents: {},
  firebaseConfig,
  // Full-page redirect sign-in is the robust default for a freshly-deployed
  // domain (popup can be blocked / closed under Chrome COOP before its result
  // reaches the opener). BaseAppComponent completes it via getRedirectResult().
  signInMethod: 'redirect',
};
