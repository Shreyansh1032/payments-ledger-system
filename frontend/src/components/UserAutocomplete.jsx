import { useState, useEffect, useRef } from 'react';
import api from '../utils/api';
import { Avatar } from './Avatar';
import { getPayWaveId } from '../utils/paywaveId';
import { useToast } from '../context/ToastContext';

export function UserAutocomplete({ label, value, onChange, placeholder }) {
  const [results, setResults] = useState([]);
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef(null);
  const { showToast } = useToast();

  useEffect(() => {
    if (!value) {
      setResults([]);
      return;
    }
    const timeout = setTimeout(async () => {
      try {
        const { data } = await api.get('/users/search', { params: { q: value } });
        setResults(data.users || []);
        setIsOpen(true);
      } catch {
        setResults([]);
      }
    }, 250);
    return () => clearTimeout(timeout);
  }, [value]);

  useEffect(() => {
    function handleClickOutside(e) {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  function handleSelect(user) {
    onChange(user.username);
    setIsOpen(false);
    setResults([]);
  }

  async function handleFavorite(e, user) {
    e.stopPropagation();
    try {
      await api.post('/favorites', { username: user.username });
      showToast(`Added ${user.first_name} to favorites`);
    } catch (err) {
      showToast(err.response?.data?.error || 'Could not add favorite', 'error');
    }
  }

  return (
    <div className="relative" ref={containerRef}>
      {label && <div className="text-sm font-medium text-left py-2 text-ink/80">{label}</div>}
      <input
        type="text"
        value={value}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        onFocus={() => value && setIsOpen(true)}
        className="w-full px-3 py-2 border border-ink/10 bg-white/70 dark:bg-black/20 dark:border-white/10 rounded-xl text-sm placeholder:text-muted focus:outline-none focus:ring-2 focus:ring-rose/30 focus:border-rose/40 transition"
      />
      {isOpen && results.length > 0 && (
        <div className="absolute z-10 mt-1 w-full glass-panel rounded-xl overflow-hidden max-h-60 overflow-y-auto">
          {results.map((u) => (
            <div
              key={u.username}
              onClick={() => handleSelect(u)}
              className="w-full flex items-center gap-2 px-3 py-2 hover:bg-white/70 dark:hover:bg-white/5 transition cursor-pointer"
            >
              <Avatar name={u.first_name} size={28} />
              <div className="min-w-0 flex-1">
                <div className="text-sm text-ink truncate">{u.first_name} {u.last_name}</div>
                <div className="text-xs text-muted truncate">{getPayWaveId(u.username)}</div>
              </div>
              <button
                type="button"
                onClick={(e) => handleFavorite(e, u)}
                className="text-muted hover:text-rose transition text-lg shrink-0"
                title="Add to favorites"
              >
                ☆
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
