import process from "node:process";

function required(environment, name) {
  const value = environment[name];
  if (typeof value !== "string" || value.trim() === "")
    throw new Error(`Cognito token helper requires ${name}`);
  return value.trim();
}

function shellQuote(value) {
  return `'${value.replaceAll("'", "'\\\"'\\\"'")}'`;
}

async function cognitoRequest(region, target, body, fetchImpl) {
  const response = await fetchImpl(`https://cognito-idp.${region}.amazonaws.com/`, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-amz-json-1.1",
      "X-Amz-Target": `AWSCognitoIdentityProviderService.${target}`,
    },
    body: JSON.stringify(body),
  });
  const result = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error("Cognito authentication failed; verify local credentials and challenge inputs");
  return result;
}

export async function acquireAccessToken(environment = process.env, fetchImpl = fetch) {
  const region = required(environment, "COGNITO_REGION");
  const clientId = required(environment, "COGNITO_CLIENT_ID");
  const username = required(environment, "COGNITO_USERNAME");
  const password = required(environment, "COGNITO_PASSWORD");
  let result = await cognitoRequest(region, "InitiateAuth", {
    AuthFlow: "USER_PASSWORD_AUTH",
    ClientId: clientId,
    AuthParameters: { USERNAME: username, PASSWORD: password },
  }, fetchImpl);

  if (result.ChallengeName === "SOFTWARE_TOKEN_MFA") {
    const code = required(environment, "COGNITO_MFA_CODE");
    result = await cognitoRequest(region, "RespondToAuthChallenge", {
      ChallengeName: "SOFTWARE_TOKEN_MFA",
      ClientId: clientId,
      Session: required({ COGNITO_SESSION: result.Session }, "COGNITO_SESSION"),
      ChallengeResponses: { USERNAME: username, SOFTWARE_TOKEN_MFA_CODE: code },
    }, fetchImpl);
  }

  const token = result?.AuthenticationResult?.AccessToken;
  if (typeof token !== "string" || token === "")
    throw new Error("Cognito did not return an access token; complete the required challenge locally");
  return token;
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  const token = await acquireAccessToken();
  process.stdout.write(`export STAGING_SMOKE_ACCESS_TOKEN=${shellQuote(token)}\n`);
}
