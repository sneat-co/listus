export const listusLocalApiBaseUrl =
  'https://sneat-api.dev.localhost:4300/v0/';

export function getListusApiBaseUrl(
  useFirebaseEmulators: boolean,
): string | undefined {
  return useFirebaseEmulators ? listusLocalApiBaseUrl : undefined;
}
