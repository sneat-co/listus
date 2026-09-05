import { listShortID } from './list-id';

describe('listShortID', () => {
  it('extracts a short ID after validating its expected type', () => {
    expect(listShortID('buy!groceries', 'buy')).toBe('groceries');
    expect(listShortID('do!tasks', 'do')).toBe('tasks');
  });

  it.each(['groceries', '!groceries', 'buy!', 'buy!one!two'])(
    'rejects malformed canonical ID %s',
    (listID) => expect(() => listShortID(listID)).toThrow(),
  );

  it('rejects a different expected type', () => {
    expect(() => listShortID('buy!groceries', 'do')).toThrow(
      'expected do',
    );
  });
});
