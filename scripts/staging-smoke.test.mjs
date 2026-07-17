import assert from "node:assert/strict";
import test from "node:test";
import { accessToken, runSmoke, validateSmokeStage } from "./staging-smoke.mjs";

const outputs = {
  apiUrl: "https://api.example.com",
  userPoolId: "us-east-1_example",
  clientId: "client",
};

test("rejects production stages without cloud access", () => {
  assert.throws(() => validateSmokeStage("PROD-uction"), /refuses production/);
});

test("uses a locally acquired access token without invoking Cognito", () => {
  assert.equal(
    accessToken(outputs, { STAGING_SMOKE_ACCESS_TOKEN: "never-logged" }, () => {
      throw new Error("Cognito should not be invoked");
    }),
    "never-logged",
  );
});

test("requires health envelope, gateway 401, and valid-token acceptance on one read-only route", async () => {
  const requests = [];
  const fetchImpl = async (url, options = {}) => {
    requests.push({ url, options });
    if (url.endsWith("/api/health"))
      return new Response(
        JSON.stringify({ status: "success", data: { status: "healthy" } }),
        { status: 200 },
      );
    if (options.headers === undefined) return new Response("", { status: 401 });
    return new Response(JSON.stringify({ status: "success", data: [] }), {
      status: 200,
    });
  };
  await runSmoke({
    stage: "staging",
    outputs,
    token: "never-logged",
    fetchImpl,
  });
  assert.equal(requests[1].url, requests[2].url);
  assert.equal(requests[1].options.headers, undefined);
  assert.match(requests[2].options.headers.Authorization, /^Bearer /);
});
