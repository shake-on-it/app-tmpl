import ApiClient, { Auth } from './base_client';
import { HttpMethods } from './types';

import { toUser, toVersion, ClientApi, User } from '../types';

export const paths = (baseUrl: string) => ({
  adminApi: (path: string) => `${baseUrl}/api/admin/v1${path}`,
  privateApi: (path: string) => `${baseUrl}/api/private/v1${path}`,
});

class Client extends ApiClient<ClientApi, User> {
  constructor(setUser: (user?: User) => void) {
    super('http://localhost:5050', setUser);
  }

  private get adminApiV1() {
    return (path: string) => `${this.baseUrl}/api/admin/v1${path}`;
  }
  private get privateApiV1() {
    return (path: string) => `${this.baseUrl}/api/private/v1${path}`;
  }

  protected get _auth(): Auth<User> {
    return {
      login: (username, password) =>
        this.do(paths(this.baseUrl).adminApi('/user/session'), {
          method: HttpMethods.Post,
          body: { username, password },
        }),
      logout: () => this.do(this.adminApiV1('/user/session'), { method: HttpMethods.Delete }),
      refresh: () => this.do(this.adminApiV1('/user/session'), { method: HttpMethods.Put, noRefresh: true }),
      whoami: () => this.do(this.adminApiV1('/user')).then(toUser),
    };
  }

  get api(): ClientApi {
    return {
      health: () => this.do(this.privateApiV1('/health'), { noAuth: true }),
      version: () => this.do(this.privateApiV1('/version'), { noAuth: true }).then(toVersion),
      errors: () => ({
        json: () => ({
          basic: () => this.do(this.privateApiV1('/errors/json/basic'), { noAuth: true }),
          complete: () => this.do(this.privateApiV1('/errors/json/complete'), { noAuth: true }),
        }),
        payload: () => this.do(this.privateApiV1('/errors/payload'), { noAuth: true }),
        text: () => this.do(this.privateApiV1('/errors/text'), { noAuth: true }),
      }),
    };
  }
}

export default Client;
