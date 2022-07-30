import React, { useEffect } from 'react';
import { useErrorHandler } from 'react-error-boundary';

import { useClient } from '../../client';

export function Bomb({ explode }: { explode: boolean }) {
  if (explode) {
    throw new Error('a bomb went off!');
  }
  return <>the bomb never went off</>;
}

export function ResponseBomb() {
  const { client } = useClient();
  const errorHandler = useErrorHandler();
  useEffect(() => {
    client.api
      .errors()
      .text()
      .catch(errorHandler);
  }, []);
  return <>response bomb</>;
}

export function PayloadBomb() {
  const { client } = useClient();
  const errorHandler = useErrorHandler();
  useEffect(() => {
    client.api
      .errors()
      .payload()
      .catch(errorHandler);
  }, []);
  return <>payload bomb</>;
}

export function ServerBomb({ complete }: { complete?: boolean }) {
  const { client } = useClient();
  const errorHandler = useErrorHandler();
  useEffect(() => {
    if (complete) {
      client.api
        .errors()
        .json()
        .complete()
        .catch(errorHandler);
      return;
    }
    client.api
      .errors()
      .json()
      .basic()
      .catch(errorHandler);
  }, []);
  return <>server bomb</>;
}
