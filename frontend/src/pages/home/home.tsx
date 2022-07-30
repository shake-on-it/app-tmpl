import React from 'react';
import { keyframes } from '@emotion/react';
import styled from '@emotion/styled';

import logo from './logo.svg';

const StyledApp = styled.article`
  text-align: center;
`;

const StyledHeader = styled.section`
  background-color: #282c34;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  font-size: calc(10px + 2vmin);
  color: white;
`;

const appLogoSpin = keyframes`
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
`;

const StyledLogo = styled.img`
  height: 40vmin;
  pointer-events: none;
  @media (prefers-reduced-motion: no-preference) {
    animation: ${appLogoSpin} infinite 20s linear;
  }
`;

const StyledLink = styled.a`
  color: #61dafb;
`;

function HomePage() {
  return (
    <StyledApp>
      <StyledHeader>
        <StyledLogo src={logo} alt="logo" />
        <p>
          Edit <code>src/App.tsx</code> and save to reload.
        </p>
        <StyledLink href="https://reactjs.org" target="_blank" rel="noopener noreferrer">
          Learn React
        </StyledLink>
      </StyledHeader>
    </StyledApp>
  );
}

export default HomePage;
