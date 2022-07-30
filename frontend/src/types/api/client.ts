import { Version } from './server';

export interface ClientApi {
  health: () => Promise<void>;
  version: () => Promise<Version>;
  errors: () => ({
    json: () => ({
      basic: () => Promise<void>;
      complete: () => Promise<void>;
    });
    payload: () => Promise<void>;
    text: () => Promise<void>;
  });
}
