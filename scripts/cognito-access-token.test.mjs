import assert from "node:assert/strict";
import test from "node:test";
import { acquireAccessToken } from "./cognito-access-token.mjs";

const environment = {
  COGNITO_REGION: "us-east-1",
  COGNITO_CLIENT_ID: "direct-client",
  COGNITO_USERNAME: "operator@example.com",
  COGNITO_PASSWORD: "never-logged",
  COGNITO_MFA_CODE: "123456",
};

test("acquires an access token after a software TOTP challenge", async () => {
  const requests = [];
  const fetchImpl = async (_, options) => {
    requests.push(JSON.parse(options.body));
    if (requests.length === 1)
      return new Response(JSON.stringify({ ChallengeName: "SOFTWARE_TOKEN_MFA", Session: "transient" }), { status: 200 });
    return new Response(JSON.stringify({ AuthenticationResult: { AccessToken: "never-logged" } }), { status: 200 });
  };
  assert.equal(await acquireAccessToken(environment, fetchImpl), "never-logged");
  assert.equal(requests[1].ChallengeResponses.SOFTWARE_TOKEN_MFA_CODE, "123456");
});

test("requires a local MFA code only when Cognito requests it", async () => {
  await assert.rejects(
    acquireAccessToken({ ...environment, COGNITO_MFA_CODE: "" }, async () => new Response(JSON.stringify({ ChallengeName: "SOFTWARE_TOKEN_MFA", Session: "transient" }), { status: 200 })),
    /COGNITO_MFA_CODE/,
  );
});
