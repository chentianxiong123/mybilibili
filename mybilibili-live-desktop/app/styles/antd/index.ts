import antdNightTheme from './night-theme.lazy.less';
import antdDayTheme from './day-theme.lazy.less';

const themes = {
  ['night-theme']: antdNightTheme,
  ['day-theme']: antdDayTheme,
};

export type Theme = keyof typeof themes;

export default themes;
