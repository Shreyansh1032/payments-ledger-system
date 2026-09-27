import { useState } from 'react';
import { getStoredTheme, applyTheme } from '../utils/theme';

export function ThemeToggle() {
  const [theme, setTheme] = useState(getStoredTheme());

  function toggle() {
    const next = theme === 'dark' ? 'light' : 'dark';
    applyTheme(next);
    setTheme(next);
  }

  return (
    <button onClick={toggle} className="text-sm text-muted hover:text-rose transition shrink-0">
      {theme === 'dark' ? 'Light mode' : 'Dark mode'}
    </button>
  );
}
