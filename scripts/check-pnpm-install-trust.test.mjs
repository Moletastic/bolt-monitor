import assert from "node:assert/strict";
import test from "node:test";
import { checkPnpmInstallTrust } from "./check-pnpm-install-trust.mjs";

const files = {
  "apps/dashboard/pnpm-lock.yaml": "lockfileVersion: 9",
  "apps/dashboard/.npmrc":
    "onlyBuiltDependencies[]=sharp\nonlyBuiltDependencies[]=unrs-resolver\nonlyBuiltDependencies[]=fsevents",
  "infra/pnpm-lock.yaml": "lockfileVersion: 9",
  "infra/.npmrc": "onlyBuiltDependencies[]=esbuild",
};

test("accepts committed lockfiles and explicit package-root allowlists", () => {
  assert.deepEqual(
    checkPnpmInstallTrust((path) => files[path]),
    [],
  );
});

test("reports the package root and missing reviewed dependency", () => {
  const broken = { ...files, "infra/.npmrc": "" };
  assert.match(
    checkPnpmInstallTrust((path) => broken[path]).join("\n"),
    /infra\/.npmrc.*esbuild/,
  );
});
