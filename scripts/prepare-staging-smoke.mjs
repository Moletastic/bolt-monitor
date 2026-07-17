import process from "node:process";
import {
  confirmationFor,
  loadDeploymentTarget,
  normalizeStageName,
} from "../infra/deployment-target.ts";

function requireValue(environment, name) {
  const value = environment[name];
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(
      `staging smoke requires ${name}; configure it in your local environment`,
    );
  }
  return value.trim();
}

export function prepareStagingSmoke(environment = process.env) {
  const stage = requireValue(environment, "SST_STAGE");
  requireValue(environment, "SST_TARGET_CONFIG");
  if (["prod", "production"].includes(normalizeStageName(stage))) {
    throw new Error(`staging smoke refuses production stage: ${stage}`);
  }
  const target = loadDeploymentTarget(stage, environment);
  if (target.lifecycle !== "persistent") {
    throw new Error(
      "staging smoke requires the declared persistent staging lifecycle; ephemeral smoke is not configured",
    );
  }
  return { target, confirmation: confirmationFor(target, "deploy") };
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  prepareStagingSmoke();
  console.log("staging smoke target configuration validated");
}
