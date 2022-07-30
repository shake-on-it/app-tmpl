import React, { createContext, useContext, useEffect, useState } from 'react';

import { Err, User, Version } from '../types';

import Client from './client';

interface Props {
  client: Client;
  healthy: boolean;
  version: Version;
  initialized: boolean;
  user?: User;
  ackErr: () => Err | undefined;
  addErr: (err: Err) => void;
}

const ClientContext = createContext<Props>(null as any);

function ClientProvider({ children }: { children: React.ReactNode }) {
  const [healthy, setHealthy] = useState(false);
  const [version, setVersion] = useState<Version>({
    env: '',
    lastCommit: '',
    buildTime: new Date(),
    time: new Date(),
  });

  const [initialized, setInitialized] = useState(false);
  const [user, setUser] = useState<User>();

  const [errs, setErrs] = useState<Err[]>([]);

  const addErr = (err: Err) => setErrs(errs.concat(err));

  const ackErr = () => {
    if (!errs.length) {
      return;
    }
    const [err, ...newErrs] = errs;
    setErrs(newErrs);
    return err;
  };

  const client = new Client(setUser);

  useEffect(() => {
    client.api
      .health()
      .then(() => {
        setHealthy(true);
        return client.api.version();
      })
      .then(setVersion)
      .catch(addErr);
  }, []);

  useEffect(() => {
    if (healthy) {
      client.auth
        .whoami()
        .catch(addErr)
        .finally(() => setInitialized(true));
    }
  }, [healthy]);

  return (
    <ClientContext.Provider
      value={{
        client,
        healthy,
        version,
        initialized,
        user,
        addErr,
        ackErr,
      }}
    >
      {children}
    </ClientContext.Provider>
  );
}

export default ClientProvider;

export function useClient() {
  return useContext(ClientContext);
}
