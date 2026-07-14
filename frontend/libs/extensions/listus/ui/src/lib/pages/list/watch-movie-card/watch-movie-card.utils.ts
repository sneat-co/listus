import { IWatchWith } from '@sneat/extension-listus-contract';

// Pure helpers for rendering a watch-list movie card (watch-movie-card.component.html).
// Kept side-effect-free & unit-tested independently of the component
// (see watch-movie-card.utils.spec.ts).

const OVERVIEW_SNIPPET_LENGTH = 120;

// Truncates a movie overview to a short snippet for the card, breaking on a
// word boundary where possible so we don't cut mid-word.
export function truncateOverview(
  overview?: string,
  maxLength = OVERVIEW_SNIPPET_LENGTH,
): string | undefined {
  if (!overview) {
    return undefined;
  }
  const trimmed = overview.trim();
  if (trimmed.length <= maxLength) {
    return trimmed;
  }
  const cut = trimmed.slice(0, maxLength);
  const lastSpace = cut.lastIndexOf(' ');
  const snippet = lastSpace > 0 ? cut.slice(0, lastSpace) : cut;
  return `${snippet.trimEnd()}…`;
}

// Joins the top cast member names into a single display line.
export function formatCastLine(cast?: string[]): string | undefined {
  if (!cast?.length) {
    return undefined;
  }
  return cast.join(', ');
}

// Builds the "watch with" chip label (alone/family/space/contact).
export function formatWatchWithLabel(watchWith?: IWatchWith): string | undefined {
  if (!watchWith) {
    return undefined;
  }
  switch (watchWith.mode) {
    case 'alone':
      return 'Alone';
    case 'space':
      return watchWith.title ? `With ${watchWith.title}` : 'With this space';
    case 'contact':
      return watchWith.title ? `With ${watchWith.title}` : 'With a contact';
    default:
      return undefined;
  }
}

// Builds the YouTube "watch" URL for a trailer key.
export function youTubeWatchUrl(trailerYouTubeKey: string): string {
  return `https://www.youtube.com/watch?v=${trailerYouTubeKey}`;
}
