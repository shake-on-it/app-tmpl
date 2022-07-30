import React from 'react';
import { useNavigate, Link } from 'react-router-dom';
import styled from '@emotion/styled';

import { useClient } from '../../client';

import { ReactComponent as Logo } from './logo.svg';

const StyledNav = styled.nav`
  flex: 1;
  display: flex;
  align-items: center;
`;

const StyledNavLink = styled(Link)`
  flex: 1;
  text-decoration: none;
  color: blue;
`;

const StyledUserInfo = styled.div`
  flex-basis: 180px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  justify-content: center;
`;

const StyledUserMessage = styled.p`
  margin-bottom: 0.25rem;
`;

function Navbar() {
  const { client, user } = useClient();
  const navigate = useNavigate();
  return (
    <>
      <Link to="/" style={{ margin: 'auto', padding: '0 0.5rem', flexBasis: '180px' }}>
        <Logo />
      </Link>
      <StyledNav>
        <StyledNavLink to="/sportsbook">Sportsbook</StyledNavLink>
        <StyledNavLink to="/eggcorn">Nowhere</StyledNavLink>
        <StyledNavLink to="/bombs/javascript">Bomb</StyledNavLink>
        <StyledNavLink to="/bombs/server/basic">Server Bomb (Basic)</StyledNavLink>
        <StyledNavLink to="/bombs/server/complete">Server Bomb (Complete)</StyledNavLink>
        <StyledNavLink to="/bombs/payload">Payload Bomb</StyledNavLink>
        <StyledNavLink to="/bombs/response">Response Bomb</StyledNavLink>
      </StyledNav>
      <StyledUserInfo>
        {!user && (
          <p style={{ marginBottom: '0.25rem' }}>
            You are not <StyledNavLink to="/login">logged in.</StyledNavLink>
          </p>
        )}
        {user && (
          <>
            <StyledUserMessage>Welcome {user.name}!</StyledUserMessage>
            <button onClick={() => client.auth.logout().then(() => navigate('/'))}>Sign out</button>
          </>
        )}
      </StyledUserInfo>
    </>
  );
}

export default Navbar;
