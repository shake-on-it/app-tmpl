const theme = {
  components: {
    navbar: { height: '60px' },
  },
  palette: {
    document: {
      background: { base: 'white' },
      text: { base: 'black' },
    },
  },
  typography: {
    fontAliases: {
      heading: 'Optima, serif',
      subheading: 'Brush Script MT, cursive',
      text: 'Gill Sans, sans-serif',
      code: 'Lucida Console, monospace',
    },
    fontWeights: {
      normal: 400,
      bold: 700,
    },
    fontSizes: {
      smallest: '10px',
      smaller: '12px',
      base: '14px',
    },
  },
};

export type Theme = typeof theme;
export default theme;
