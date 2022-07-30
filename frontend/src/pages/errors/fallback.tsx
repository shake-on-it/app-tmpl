import React from 'react';
import { FallbackProps } from 'react-error-boundary';

import { PayloadError, ResponseError, ServerError } from '../../types';

function ServerErrorDisplay({ error }: { error: ServerError }) {
  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <h3>{error.name}</h3>
        <p>Status: {error.status}</p>
      </div>
      <ErrorMessage message={error.message} />
      <hr />
      {error.data && (
        <pre style={{ overflowX: 'auto' }}>
          <code>{JSON.stringify(error.data, null, 2)}</code>
        </pre>
      )}
      <p style={{ textAlign: 'right' }}>Request ID: {error.requestId}</p>
    </>
  );
}

function PayloadErrorDisplay({ error }: { error: PayloadError }) {
  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <h3>{error.name}</h3>
        <p>Status: {error.status}</p>
      </div>
      <ErrorMessage message={error.message} />
      <hr />
      <pre style={{ overflowX: 'auto' }}>
        <code>{JSON.stringify(error.payload, null, 2)}</code>
      </pre>
    </>
  );
}

function ResponseErrorDisplay({ error }: { error: ResponseError }) {
  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <h3>{error.name}</h3>
        <p>Status: {(error as any).status}</p>
      </div>
      <ErrorMessage message={error.message} />
    </>
  );
}

function ErrorDisplay({ error }: { error: Error }) {
  return (
    <>
      <h2>{error.name}</h2>
      <ErrorMessage message={error.message} />
    </>
  );
}

function ErrorMessage({ message }: { message: string }) {
  return (
    <pre style={{ whiteSpace: 'pre-wrap', overflowX: 'auto' }}>
      <code>{message}</code>
    </pre>
  );
}

function ErrorFallback({ error, resetErrorBoundary }: FallbackProps) {
  return (
    <article style={{ width: '50%', margin: 'auto', display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
      <h1>OOPS!</h1>
      <p>Things aren't going so well over here</p>
      <p>I'd suggest you try again later</p>
      <section style={{ width: '50%', marginBottom: '1rem', padding: '0 1rem', border: '1px solid' }}>
        {error instanceof ServerError ? (
          <ServerErrorDisplay error={error as any} />
        ) : error instanceof PayloadError ? (
          <PayloadErrorDisplay error={error} />
        ) : error instanceof ResponseError ? (
          <ResponseErrorDisplay error={error} />
        ) : (
          <ErrorDisplay error={error} />
        )}
      </section>
      <nav>
        <button onClick={resetErrorBoundary}>Back to Safety</button>
      </nav>
    </article>
  );
}

export default ErrorFallback;
