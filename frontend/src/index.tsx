import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';

import { Global, ThemeProvider } from '@emotion/react';

import App from './app';
import ClientProvider from './client';
import reportWebVitals from './reportWebVitals';
import styles from './styles';
import theme from './theme';

const rootId = 'root';

const root = ReactDOM.createRoot(document.getElementById(rootId) as HTMLElement);
root.render(
  <React.StrictMode>
    <BrowserRouter>
      <ClientProvider>
        <ThemeProvider theme={theme}>
          <Global
            styles={[
              styles.meyers(),
              styles.borderBox(),
              styles.fonts(theme),
              styles.body(theme),
              { ['#' + rootId]: { minHeight: '100vh' } },
            ]}
          />
          <App />
        </ThemeProvider>
      </ClientProvider>
    </BrowserRouter>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
