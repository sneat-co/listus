import { builtInListGroups, listGroupsFromBriefs } from './built-in-lists';

describe('builtInListGroups', () => {
  it.each(['personal', 'family'] as const)('provides starter lists for a %s space', (spaceType) => {
    expect(builtInListGroups(spaceType)).toEqual(expect.arrayContaining([
      expect.objectContaining({ type: 'buy' }),
      expect.objectContaining({ type: 'do' }),
    ]));
  });

  it('does not treat unrelated group spaces as personal', () => {
    expect(builtInListGroups('group')).toEqual([]);
  });
});

describe('listGroupsFromBriefs', () => {
  const created = { createdAt: { seconds: 1, nanoseconds: 0 }, createdBy: 'u1' };

  it('groups canonical extension-owned briefs and keeps short route IDs', () => {
    expect(
      listGroupsFromBriefs({
        'do!payments': { ...created, type: 'do', title: 'Payments' },
        'buy!food': { ...created, type: 'buy', title: 'Food' },
      }),
    ).toEqual([
      {
        id: 'buy',
        type: 'buy',
        lists: [
          expect.objectContaining({ id: 'food', shortId: 'food', title: 'Food' }),
        ],
      },
      {
        id: 'do',
        type: 'do',
        lists: [
          expect.objectContaining({
            id: 'payments',
            shortId: 'payments',
            title: 'Payments',
          }),
        ],
      },
    ]);
  });

  it('ignores malformed or mismatched qualified IDs', () => {
    expect(
      listGroupsFromBriefs({
        missing: { ...created, type: 'do', title: 'Missing separator' },
        'buy!wrong': { ...created, type: 'do', title: 'Wrong type' },
      }),
    ).toEqual([]);
  });
});
