import {
  formatCastLine,
  formatWatchWithLabel,
  truncateOverview,
  youTubeWatchUrl,
} from './watch-movie-card.utils';

describe('truncateOverview', () => {
  it('returns undefined when there is no overview', () => {
    expect(truncateOverview(undefined)).toBeUndefined();
  });

  it('returns the overview unchanged when shorter than the limit', () => {
    expect(truncateOverview('A short overview.')).toBe('A short overview.');
  });

  it('truncates a long overview on a word boundary and appends an ellipsis', () => {
    const overview =
      'A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a CEO.';
    const result = truncateOverview(overview, 40);
    expect(result?.length).toBeLessThanOrEqual(41); // 40 + ellipsis char
    expect(result?.endsWith('…')).toBe(true);
    expect(overview.startsWith(result?.slice(0, -1) ?? '')).toBe(true);
  });
});

describe('formatCastLine', () => {
  it('returns undefined when there is no cast', () => {
    expect(formatCastLine(undefined)).toBeUndefined();
    expect(formatCastLine([])).toBeUndefined();
  });

  it('joins cast member names with a comma', () => {
    expect(formatCastLine(['Leonardo DiCaprio', 'Joseph Gordon-Levitt'])).toBe(
      'Leonardo DiCaprio, Joseph Gordon-Levitt',
    );
  });
});

describe('formatWatchWithLabel', () => {
  it('returns undefined when there is no watchWith', () => {
    expect(formatWatchWithLabel(undefined)).toBeUndefined();
  });

  it('labels alone mode', () => {
    expect(formatWatchWithLabel({ mode: 'alone' })).toBe('Alone');
  });

  it('labels space mode using the denormalized title', () => {
    expect(
      formatWatchWithLabel({ mode: 'space', ref: 'space1', title: 'Family' }),
    ).toBe('With Family');
  });

  it('falls back to a generic label for space mode with no title', () => {
    expect(formatWatchWithLabel({ mode: 'space', ref: 'space1' })).toBe(
      'With this space',
    );
  });

  it('labels contact mode using the denormalized title', () => {
    expect(
      formatWatchWithLabel({ mode: 'contact', ref: 'c1', title: 'Alice' }),
    ).toBe('With Alice');
  });

  it('falls back to a generic label for contact mode with no title', () => {
    expect(formatWatchWithLabel({ mode: 'contact', ref: 'c1' })).toBe(
      'With a contact',
    );
  });
});

describe('youTubeWatchUrl', () => {
  it('builds a youtube watch URL from a trailer key', () => {
    expect(youTubeWatchUrl('abc123')).toBe(
      'https://www.youtube.com/watch?v=abc123',
    );
  });
});
