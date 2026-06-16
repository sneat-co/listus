import { appSpecificConfig, emulatorEnvironmentConfig } from '@sneat/app';
import { IEnvironmentConfig } from '@sneat/core';

// Mirrors the sneat-app environment: app-specific config layered over the
// Firebase emulator base. Swap the base for a production config when listus
// gets its own Firebase project.
export const listusAppEnvironmentConfig: IEnvironmentConfig =
  appSpecificConfig(emulatorEnvironmentConfig);
