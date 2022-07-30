import React, { useState } from 'react';

import { useClient } from '../../client';

function SportsbookPage() {
  const { client } = useClient();
  const [err, setErr] = useState('');
  return (
    <article>
      <h1>Sportsbook Page</h1>
      <button
        onClick={() =>
          client.auth
            .whoami()
            .then(() => console.log("that's a bingo!"))
            .catch(err => setErr(String(err)))
        }
      >
        Submit Request
      </button>
      <p>Error: {err}</p>
    </article>
  );
}

export default SportsbookPage;
