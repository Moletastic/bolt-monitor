import { execFileSync } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import process from 'node:process';
import { loadDeploymentTarget } from '../infra/deployment-target.ts';

function requireStage(environment) {
  const target = loadDeploymentTarget(environment.SST_STAGE, environment);
  return {
    target,
    parameterName: `/${target.service}/${target.stage}/auth/aes-256-gcm`,
  };
}

export function rotateAuthKey(environment = process.env, execute = execFileSync) {
  const { target, parameterName } = requireStage(environment);
  // Key exists only in process memory and the AWS CLI invocation boundary.
  const key = randomBytes(32).toString('base64url');
  execute(
    'aws',
    [
      'ssm',
      'put-parameter',
      '--name',
      parameterName,
      '--type',
      'SecureString',
      '--value',
      key,
      '--overwrite',
      '--output',
      'json',
    ],
    { env: environment, stdio: 'pipe' }
  );
  return { parameterName, stage: target.stage };
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  try {
    rotateAuthKey();
  } catch {
    console.error('auth key rotation failed')
    process.exitCode = 1
  }
}
