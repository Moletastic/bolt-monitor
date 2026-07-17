import { createRoute } from './route-record.mjs';

function lineNumber(source, index) {
  return source.slice(0, index).split('\n').length;
}

function closingBrace(source, openingIndex) {
  let depth = 0;
  let quoted = false;
  let escaped = false;
  for (let index = openingIndex; index < source.length; index += 1) {
    const character = source[index];
    if (quoted) {
      if (!escaped && character === '"') quoted = false;
      escaped = !escaped && character === '\\';
      continue;
    }
    if (character === '"') quoted = true;
    if (character === '{') depth += 1;
    if (character === '}') {
      depth -= 1;
      if (depth === 0) return index;
    }
  }
  return -1;
}

function quotedValues(value) {
  const values = [];
  let index = 0;
  while (index < value.length) {
    while (/\s|,/.test(value[index] ?? '')) index += 1;
    if (index === value.length) break;
    if (value[index] !== '"') return undefined;
    const end = (() => {
      let escaped = false;
      for (let cursor = index + 1; cursor < value.length; cursor += 1) {
        if (!escaped && value[cursor] === '"') return cursor;
        escaped = !escaped && value[cursor] === '\\';
      }
      return -1;
    })();
    if (end < 0) return undefined;
    values.push(JSON.parse(value.slice(index, end + 1)));
    index = end + 1;
  }
  return values;
}

export function extractHandlerRouteInventory(source, filePath = 'services/monitor-api/routes.go') {
  const declaration = source.indexOf('var monitorRoutes = []routeDefinition{');
  if (declaration < 0) return { routes: [], diagnostics: [`${filePath}: monitorRoutes inventory declaration not found`] };
  const opening = source.indexOf('{', declaration);
  const closing = closingBrace(source, opening);
  if (closing < 0) return { routes: [], diagnostics: [`${filePath}:${lineNumber(source, opening)}: monitorRoutes inventory is not closed`] };

  const routes = [];
  const diagnostics = [];
  const inventory = source.slice(opening + 1, closing);
  for (let index = 0; index < inventory.length; index += 1) {
    if (inventory[index] !== '{') continue;
    const entryEnd = closingBrace(inventory, index);
    if (entryEnd < 0) {
      diagnostics.push(`${filePath}:${lineNumber(source, opening + 1 + index)}: unsupported dynamic monitorRoutes entry; use a literal method, path, and handler`);
      break;
    }
    const values = quotedValues(inventory.slice(index + 1, entryEnd));
    if (!values || values.length !== 3) {
      diagnostics.push(`${filePath}:${lineNumber(source, opening + 1 + index)}: unsupported dynamic monitorRoutes entry; use a literal method, path, and handler`);
    } else {
      try {
        routes.push(createRoute({ method: values[0], path: values[1], target: values[2], source: `${filePath}:${lineNumber(source, opening + 1 + index)}` }));
      } catch (error) {
        diagnostics.push(`${filePath}:${lineNumber(source, opening + 1 + index)}: invalid monitorRoutes entry: ${error.message}`);
      }
    }
    index = entryEnd;
  }
  return { routes, diagnostics };
}
