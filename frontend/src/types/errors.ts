import { ApiErr } from './api';

export class ResponseError extends Error {
  constructor(message: string, public status: number) {
    super(message);
    this.name = 'ResponseError';
  }
}

export class ServerError extends ResponseError {
  public requestId: string;
  public data?: object;

  constructor(err: ApiErr, status: number) {
    super(err.msg, status);
    this.name = (err.code || 'Unknown').split('_').map(n => n.charAt(0).toUpperCase() + n.substring(1)) + 'Error';
    this.requestId = err.request_id;
    this.data = err.data;
  }
}

export class PayloadError extends ResponseError {
  constructor(public payload: any, status: number) {
    super('an error response was returned', status);
    this.name = 'ResponseError';
  }
}
