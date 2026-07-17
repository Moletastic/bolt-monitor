import { readFileSync } from "node:fs";

const root = new URL("..", import.meta.url);

const packageRoots = [
  {
    path: "apps/dashboard",
    trustedPackages: ["sharp", "unrs-resolver", "fsevents"],
  },
  { path: "infra", trustedPackages: ["esbuild"] },
];

function source(path) {
  return readFileSync(new URL(path, root), "utf8");
}

function allowlist(npmrc) {
  return [...npmrc.matchAll(/^onlyBuiltDependencies\[\]=(.+)$/gm)].map(
    (match) => match[1],
  );
}

export function checkPnpmInstallTrust(read = source) {
  const errors = [];
  for (const packageRoot of packageRoots) {
    const lockfile = `${packageRoot.path}/pnpm-lock.yaml`;
    const npmrc = `${packageRoot.path}/.npmrc`;
    try {
      read(lockfile);
    } catch {
      errors.push(`${lockfile}: committed pnpm lockfile is required`);
    }
    let actual;
    try {
      actual = allowlist(read(npmrc));
    } catch {
      errors.push(
        `${npmrc}: explicit onlyBuiltDependencies allowlist is required`,
      );
      continue;
    }
    for (const dependency of packageRoot.trustedPackages) {
      if (!actual.includes(dependency)) {
        errors.push(
          `${npmrc}: add reviewed install-script dependency ${dependency} to onlyBuiltDependencies`,
        );
      }
    }
  }
  return errors;
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  const errors = checkPnpmInstallTrust();
  for (const error of errors) console.error(`ERROR ${error}`);
  if (errors.length > 0) process.exitCode = 1;
  else
    console.log(
      "pnpm frozen-lockfile and install-script trust policy verified",
    );
}
