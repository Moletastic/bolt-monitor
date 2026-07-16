import assert from 'node:assert/strict';
import test from 'node:test';

import { checkAuthCutoverPrerequisites } from './check-auth-cutover-prerequisites.mjs';

test('auth cutover is gated on lifecycle classification, protection, cleanup, and inventory', () => {
  assert.doesNotThrow(checkAuthCutoverPrerequisites);
});
