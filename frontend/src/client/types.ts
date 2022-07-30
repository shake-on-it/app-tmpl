export const HttpMethods = {
  Get: 'GET',
  Post: 'POST',
  Put: 'PUT',
  Patch: 'PATCH',
  Delete: 'DELETE',
  Options: 'OPTIONS',
} as const;
export type HttpMethod = typeof HttpMethods[keyof typeof HttpMethods];

export const HttpHeaders = {
  Accept: 'Accept',
  Authorization: 'Authorization',
  ContentType: 'Content-Type',
} as const;
export type HttpHeader = typeof HttpHeaders[keyof typeof HttpHeaders];

export const MediaTypes = {
  Json: 'application/json',
  Text: 'text/plain',
} as const;
export type MediaType = typeof MediaTypes[keyof typeof MediaTypes];
