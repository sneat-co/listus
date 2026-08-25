import { builtInListGroups } from './built-in-lists';

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
