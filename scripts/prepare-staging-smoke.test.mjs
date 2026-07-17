import assert from "node:assert/strict";
import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { prepareStagingSmoke } from "./prepare-staging-smoke.mjs";

function environment(target) {
  const path = join(mkdtempSync(join(tmpdir(), "bolt-smoke-")), "target.json");
  writeFileSync(path, JSON.stringify({ targets: [target] }));
  return { SST_STAGE: target.stage, SST_TARGET_CONFIG: path };
}

const target = {
  stage: "staging",
  lifecycle: "persistent",
  owner: "platform",
  service: "bolt-monitor",
  accountId: "123456789012",
  region: "us-east-1",
  credentialSource: "OIDC",
  dashboardOrigin: "https://staging.example.com",
  approved: true,
};

test("accepts only a declared persistent non-production target", () => {
  assert.match(
    prepareStagingSmoke(environment(target)).confirmation,
    /^[a-f0-9]{64}$/,
  );
});

test("rejects production and ephemeral smoke targets before AWS work", () => {
  assert.throws(
    () =>
      prepareStagingSmoke({ ...environment(target), SST_STAGE: "production" }),
    /refuses production/,
  );
  assert.throws(
    () =>
      prepareStagingSmoke(
        environment({
          ...target,
          stage: "smoke-123",
          lifecycle: "ephemeral",
          approved: undefined,
          disposable: true,
          expiresAt: "2099-01-01T00:00:00Z",
        }),
      ),
    /declared persistent/,
  );
});
