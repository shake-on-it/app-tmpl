import '@emotion/react';

import { Theme as AppTheme } from '../src/theme';

declare module '@emotion/react' {
  export interface Theme extends AppTheme {}
}
