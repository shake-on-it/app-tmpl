import React from 'react';
import { Link } from 'react-router-dom';
import styled from '@emotion/styled';

import { useClient } from '../../client';

import { ReactComponent as Logo } from './logo.svg';

const StyledNav = styled.nav`
  margin: 0.5rem;
  flex: 1;
  display: flex;
  flex-direction: column;
`;

const StyledNavLink = styled(Link)`
  text-decoration: none;
  color: blue;
  flex-basis: 30px;
`;

const StyledLogo = styled.aside`
  flex: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-decoration: none;
`;

const StyledSystemInfo = styled.section`
  margin: 0.5rem;
  flex: 1;
  display: grid;
  grid-template-columns: 1fr auto;
  column-gap: 5px;
  row-gap: 5px;
`;

const StyledVersionLabel = styled.code`
  text-align: right;
  font-size: ${({ theme }) => theme.typography.fontSizes.smaller};
`;

const StyledVersionData = styled.code`
  font-size: ${({ theme }) => theme.typography.fontSizes.smaller};
`;

function Navbar() {
  const { healthy, version } = useClient();
  return (
    <>
      <StyledNav>
        <StyledNavLink to="/sportsbook">About us</StyledNavLink>
        <StyledNavLink to="/eggcorn">Terms of Service</StyledNavLink>
        <StyledNavLink to="/bomb">Responsible gamining</StyledNavLink>
        <StyledNavLink to="/bomb_server_basic">Privacy policy</StyledNavLink>
      </StyledNav>
      <StyledLogo>
        <h2 style={{ marginBottom: '0.5rem' }}>App Mantra</h2>
        <Link to="/" style={{}}>
          <Logo />
        </Link>
      </StyledLogo>
      <StyledSystemInfo>
        <StyledVersionLabel>Status:</StyledVersionLabel>
        <StyledVersionData>
          <a
            href="http://localhost:5050/api/private/v1/version"
            target="_blank"
            rel="noopener noreferrer"
            style={{ textDecoration: 'none', color: healthy ? 'green' : 'red' }}
          >
            {healthy ? 'OK' : 'Not OK'}
          </a>
        </StyledVersionData>
        <StyledVersionLabel>Env:</StyledVersionLabel>
        <StyledVersionData>{version.env}</StyledVersionData>
        <StyledVersionLabel>Version:</StyledVersionLabel>
        <StyledVersionData>
          <a
            href={'https://github.com/shake-on-it/app-tmpl/commit/' + version.lastCommit}
            target="_blank"
            rel="noopener noreferrer"
            style={{ textDecoration: 'none', color: 'blue' }}
          >
            {version.lastCommit ? version.lastCommit.substring(0, 6) : '--'}
          </a>
        </StyledVersionData>
        <StyledVersionLabel>Build time:</StyledVersionLabel>
        <StyledVersionData>{version.buildTime ? version.buildTime.toLocaleString() : '--'}</StyledVersionData>
        <StyledVersionLabel>Last refresh:</StyledVersionLabel>
        <StyledVersionData>{version.time ? version.time.toLocaleString() : '--'}</StyledVersionData>
      </StyledSystemInfo>
    </>
  );
}

export default Navbar;
