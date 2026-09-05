import { ListType } from '@sneat/extension-listus-contract';

/** Returns the route-facing short ID from a canonical `${type}!${shortID}` ID. */
export function listShortID(listID: string, expectedType?: ListType): string {
  const separator = listID.indexOf('!');
  if (separator <= 0 || separator === listID.length - 1) {
    throw new Error(`Invalid canonical list ID: ${listID}`);
  }
  const type = listID.slice(0, separator);
  const shortID = listID.slice(separator + 1);
  if (shortID.includes('!')) {
    throw new Error(`Invalid canonical list ID: ${listID}`);
  }
  if (expectedType && type !== expectedType) {
    throw new Error(
      `List ID ${listID} has type ${type}; expected ${expectedType}`,
    );
  }
  return shortID;
}
