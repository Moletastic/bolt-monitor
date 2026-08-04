import { createRoute } from './route-record.mjs';
import { readRepositoryFile, walkFiles } from './helpers.mjs';

export function extractMergedOpenSpecRoutes(specDirectory) {
  return walkFiles(specDirectory, '.md').flatMap((filePath) => {
    const source = readRepositoryFile(filePath);
    return [...source.matchAll(/`(GET|POST|PUT|PATCH|DELETE)\s+(\/api\/[^`\s?]+)(?:\?[^`]*)?`/g)].map((match) => createRoute({
      method: match[1],
      path: match[2],
      source: `${filePath}:${source.slice(0, match.index).split('\n').length}`,
    }));
  });
}
