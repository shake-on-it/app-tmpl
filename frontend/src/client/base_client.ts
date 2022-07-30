import { isApiErr, PayloadError, ResponseError, ServerError } from '../types';
import { HttpHeaders, HttpMethods, MediaTypes } from './types';

export interface Auth<USER> {
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
  whoami: () => Promise<USER>;
}

interface RequestOptions {
  method?: string;
  body?: object;
  noAuth?: true;
  noRefresh?: true;
}

export abstract class ApiClient<API, USER> {
  constructor(protected baseUrl: string, private setUser: (user?: USER) => void = () => {}) {}

  abstract get api(): API;

  protected abstract get _auth(): Auth<USER>;

  get auth(): Auth<USER> {
    return {
      login: (username, password) =>
        this._auth
          .login(username, password)
          .then(() => this._auth.whoami())
          .then(this.setUser),
      logout: () => this._auth.logout().then(() => this.setUser(undefined)),
      refresh: () =>
        this._auth
          .refresh()
          .then(() => this._auth.whoami())
          .then(this.setUser),
      whoami: () =>
        this._auth.whoami().then(user => {
          this.setUser(user);
          return user;
        }),
    };
  }

  protected async do(url: string, opts: RequestOptions = {}): Promise<any> {
    const headers = new Headers();

    let body;
    if (opts.body) {
      body = JSON.stringify(opts.body);
      headers.append(HttpHeaders.ContentType, MediaTypes.Json);
    }

    const res = await fetch(url, {
      method: opts.method || HttpMethods.Get,
      body,
      headers,
      credentials: opts.noAuth ? undefined : 'include',
    });
    const contentType = res.headers.get(HttpHeaders.ContentType);

    if (res.status === 401) {
      if (opts.noRefresh) {
        this.setUser(undefined);
      } else {
        await this.auth.refresh();
        opts.noRefresh = true;
        return this.do(url, opts);
      }
    }

    if (res.status >= 200 && res.status < 300) {
      if (res.status === 204 || !contentType) {
        return undefined;
      }
      switch (contentType) {
        case MediaTypes.Json:
          return res.json();
        case MediaTypes.Text:
          return res.text();
      }
      throw new Error(`unrecognized response content type: ${contentType}`);
    }

    switch (contentType) {
      case MediaTypes.Json: {
        let payload = await res.json();
        if (isApiErr(payload)) {
          throw new ServerError(payload, res.status);
        }
        throw new PayloadError(payload, res.status);
      }
      case MediaTypes.Text: {
        let payload = await res.text();
        throw new ResponseError(payload, res.status);
      }
    }

    throw new Error(`unrecognized response content type: ${contentType}`);
  }
}

export default ApiClient;
