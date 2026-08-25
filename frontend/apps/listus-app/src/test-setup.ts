// jsdom has no layout engine and doesn't implement `window.matchMedia`.
// Ionic 9's <ion-split-pane> calls it unconditionally from connectedCallback
// to decide whether the pane starts visible, so any spec that renders the
// app shell (<ion-app>) throws an unhandled `TypeError: window.matchMedia is
// not a function` rejection. Stub a minimal, always-non-matching
// implementation so that call succeeds instead of masking a real error.
if (typeof window !== 'undefined' && typeof window.matchMedia !== 'function') {
  window.matchMedia = (query: string): MediaQueryList =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => undefined,
      removeListener: () => undefined,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
      dispatchEvent: () => false,
    }) as MediaQueryList;
}
