import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";
import process from "node:process";
import { normalizeStageName } from "../infra/deployment-target.ts";

const defaultProtectedPath = "/api/v1/services";

function requireText(value, name) {
  if (typeof value !== "string" || value.trim() === "")
    throw new Error(`staging smoke requires ${name}`);
  return value.trim();
}

export function readSmokeOutputs(path) {
  let outputs;
  try {
    outputs = JSON.parse(readFileSync(path, "utf8"));
  } catch {
    throw new Error("staging smoke could not read structured SST outputs");
  }
  for (const name of [
    "apiUrl",
    "operatorUserPoolId",
    "directOperatorUserPoolClientId",
  ]) {
    requireText(outputs[name], `SST output ${name}`);
  }
  let apiURL;
  try {
    apiURL = new URL(outputs.apiUrl);
  } catch {
    throw new Error("staging smoke SST output apiUrl must be an absolute URL");
  }
  if (apiURL.protocol !== "https:")
    throw new Error("staging smoke SST output apiUrl must use HTTPS");
  return {
    apiUrl: apiURL.toString().replace(/\/$/, ""),
    userPoolId: outputs.operatorUserPoolId,
    clientId: outputs.directOperatorUserPoolClientId,
  };
}

export function validateSmokeStage(stage) {
  const value = requireText(stage, "SST_STAGE");
  if (["prod", "production"].includes(normalizeStageName(value)))
    throw new Error(`staging smoke refuses production stage: ${value}`);
  return value;
}

export function accessToken(
  outputs,
  environment = process.env,
  execute = execFileSync,
) {
  const suppliedToken = environment.STAGING_SMOKE_ACCESS_TOKEN;
  if (typeof suppliedToken === "string" && suppliedToken.trim() !== "")
    return suppliedToken.trim();
  const username = requireText(
    environment.STAGING_SMOKE_USERNAME,
    "STAGING_SMOKE_USERNAME protected secret",
  );
  const password = requireText(
    environment.STAGING_SMOKE_PASSWORD,
    "STAGING_SMOKE_PASSWORD protected secret",
  );
  const region = outputs.userPoolId.split("_")[0];
  let response;
  try {
    response = JSON.parse(
      execute(
        "aws",
        [
          "cognito-idp",
          "initiate-auth",
          "--region",
          region,
          "--client-id",
          outputs.clientId,
          "--auth-flow",
          "USER_PASSWORD_AUTH",
          "--auth-parameters",
          `USERNAME=${username},PASSWORD=${password}`,
          "--output",
          "json",
          "--no-cli-pager",
        ],
        { encoding: "utf8", env: environment },
      ),
    );
  } catch {
    throw new Error(
      "staging smoke could not acquire a Cognito access token; supply STAGING_SMOKE_ACCESS_TOKEN after MFA or verify local direct-client credentials",
    );
  }
  return requireText(
    response?.AuthenticationResult?.AccessToken,
    "Cognito access token",
  );
}

async function expectStatus(response, expected, name) {
  if (response.status !== expected)
    throw new Error(
      `${name}: expected HTTP ${expected}, got ${response.status}`,
    );
}

export async function runSmoke({
  stage,
  outputs,
  token,
  protectedPath = defaultProtectedPath,
  fetchImpl = fetch,
}) {
  validateSmokeStage(stage);
  if (protectedPath !== defaultProtectedPath)
    throw new Error(
      `staging smoke protected route must remain the read-only ${defaultProtectedPath} route`,
    );
  const health = await fetchImpl(`${outputs.apiUrl}/api/health`);
  await expectStatus(health, 200, "public health");
  let healthBody;
  try {
    healthBody = await health.json();
  } catch {
    throw new Error("public health: expected JSON success envelope");
  }
  if (
    healthBody?.status !== "success" ||
    healthBody?.data?.status !== "healthy"
  ) {
    throw new Error(
      "public health: expected standard healthy success envelope",
    );
  }
  const missingToken = await fetchImpl(`${outputs.apiUrl}${protectedPath}`);
  await expectStatus(missingToken, 401, "protected route without token");
  const accepted = await fetchImpl(`${outputs.apiUrl}${protectedPath}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!accepted.ok)
    throw new Error(
      `protected route with valid token: expected documented non-authentication success, got HTTP ${accepted.status}`,
    );
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  const stage = validateSmokeStage(process.env.SST_STAGE);
  const outputs = readSmokeOutputs(
    requireText(process.env.SST_OUTPUT_PATH, "SST_OUTPUT_PATH"),
  );
  const token = accessToken(outputs);
  await runSmoke({
    stage,
    outputs,
    token,
    protectedPath: defaultProtectedPath,
  });
  console.log(
    "staging smoke passed: public health, protected 401, and valid-token acceptance",
  );
}
