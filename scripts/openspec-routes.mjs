import fs from 'node:fs';
import path from 'node:path';
import { createRoute } from './route-record.mjs';

function walk(directory) {
  return fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const filePath = path.join(directory, entry.name);
    return entry.isDirectory() ? walk(filePath) : [filePath];
  });
}

export function extractMergedOpenSpecRoutes(specDirectory) {
  return walk(specDirectory).filter((filePath) => filePath.endsWith('.md')).flatMap((filePath) => {
    const source = fs.readFileSync(filePath, 'utf8');
    return [...source.matchAll(/`(GET|POST|PUT|PATCH|DELETE)\s+(\/api\/[^`\s?]+)(?:\?[^`]*)?`/g)].map((match) => createRoute({
      method: match[1],
      path: match[2],
      source: `${filePath}:${source.slice(0, match.index).split('\n').length}`,
    }));
  });
}
