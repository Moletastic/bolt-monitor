import fs from 'node:fs';
import path from 'node:path';

export const repositoryRoot = new URL('..', import.meta.url);

export function readRepositoryFile(relativePath) {
  return fs.readFileSync(new URL(relativePath, repositoryRoot), 'utf8');
}

export function walkFiles(directory, extension) {
  return fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const filePath = path.join(directory, entry.name);
    return entry.isDirectory()
      ? walkFiles(filePath, extension)
      : filePath.endsWith(extension)
        ? [filePath]
        : [];
  });
}

export function reportErrors(errors, successMessage) {
  for (const error of errors) console.error(`ERROR ${error}`);
  if (errors.length > 0) {
    process.exitCode = 1;
  } else if (successMessage !== undefined) {
    console.log(successMessage);
  }
}
