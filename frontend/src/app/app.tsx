import React, { useState } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { useNavigate, Routes, Route } from 'react-router-dom';

import styled from '@emotion/styled';

import { useClient } from '../client';

import AuthRoute from './auth_route';
import Footer from './footer';
import Navbar from './navbar';

import LoginPage from '../pages/auth/login';
import { ErrorFallback, NotFoundPage, Bomb, PayloadBomb, ResponseBomb, ServerBomb } from '../pages/errors';
import HomePage from '../pages/home';
import FAQPage from '../pages/info/faq';
import PrivacyPolicyPage from '../pages/info/privacy_policy';
import ResponsibleGamingPage from '../pages/info/responsible_gaming';
import TermsOfServicePage from '../pages/info/terms_of_service';
import SportsbookPage from '../pages/sportsbook';

const StyledHeader = styled.header`
  display: flex;
  justify-content: space-between;
  height: ${({ theme }) => theme.components.navbar.height};
  border-bottom: 1px solid;
`;

const StyledMain = styled.main`
  min-height: calc(100vh - ${({ theme }) => theme.components.navbar.height});
  min-height: -o-calc(100vh - ${({ theme }) => theme.components.navbar.height}); /* opera */
  min-height: -webkit-calc(100vh - ${({ theme }) => theme.components.navbar.height}); /* google, safari */
  min-height: -moz-calc(100vh - ${({ theme }) => theme.components.navbar.height}); /* firefox */
`;

const StyledFooter = styled.footer`
  display: flex;
  justify-content: space-between;
  border-top: 1px solid;
`;

function App() {
  const [errorReset, setErrorReset] = useState(false);
  const { initialized } = useClient();
  const navTo = useNavigate();

  if (!initialized) {
    return <>initializing...</>;
  }

  return (
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onError={(err, info) => console.error(`${err.name}: ${err.message}`, info.componentStack.toString())}
      onReset={() => {
        setErrorReset(!errorReset);
        navTo('/');
      }}
      resetKeys={[errorReset]}
    >
      <StyledHeader>
        <Navbar />
      </StyledHeader>
      <StyledMain>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route
            path="sportsbook"
            element={
              <AuthRoute>
                <SportsbookPage />
              </AuthRoute>
            }
          />

          {/* auth */}
          <Route path="login" element={<LoginPage />} />

          {/* info */}
          <Route path="faq" element={<FAQPage />} />
          <Route path="privacy_policy" element={<PrivacyPolicyPage />} />
          <Route path="responsible_gaming" element={<ResponsibleGamingPage />} />
          <Route path="terms_of_service" element={<TermsOfServicePage />} />

          {/* errors */}
          <Route path="bombs">
            <Route path="javascript" element={<Bomb explode />} />
            <Route path="server">
              <Route path="basic" element={<ServerBomb />} />
              <Route path="complete" element={<ServerBomb complete />} />
            </Route>
            <Route path="payload" element={<PayloadBomb />} />
            <Route path="response" element={<ResponseBomb />} />
          </Route>
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </StyledMain>
      <StyledFooter>
        <Footer />
      </StyledFooter>
    </ErrorBoundary>
  );
}

export default App;
