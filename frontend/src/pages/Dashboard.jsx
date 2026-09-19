import { useEffect, useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Avatar } from '../components/Avatar';
import { ErrorBanner } from '../components/ErrorBanner';
import api from '../utils/api';

export function Dashboard() {
  const [balance, setBalance] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  const username = localStorage.getItem('username') || '';
  const firstName = localStorage.getItem('first_name') || username;

  useEffect(() => {
    async function fetchBalance() {
      try {
        const { data } = await api.get('/account/balance');
        setBalance(data.balance);
      } catch (err) {
        setError(err.response?.data?.error || 'Could not load balance');
      } finally {
        setLoading(false);
      }
    }
    fetchBalance();
  }, []);

  function handleLogout() {
    const refreshToken = localStorage.getItem('refresh_token');
    api.post('/auth/logout', { refresh_token: refreshToken }).finally(() => {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('username');
      localStorage.removeItem('first_name');
      navigate('/signin');
    });
  }

  return (
    <div className="aura-backdrop min-h-screen px-4 py-8">
      <div className="max-w-md mx-auto">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <Avatar name={firstName} size={44} />
            <div>
              <div className="text-xs text-muted">Welcome back</div>
              <div className="font-display text-lg text-ink">{firstName}</div>
            </div>
          </div>
          <button onClick={handleLogout} className="text-sm text-muted hover:text-rose transition">
            Log out
          </button>
        </div>

        <ErrorBanner message={error} />

        <div className="glass-panel rounded-2xl p-8 text-center mb-6">
          <div className="text-sm text-muted mb-2">Your balance</div>
          {loading ? (
            <div className="font-display text-4xl text-ink/40">…</div>
          ) : (
            <div className="font-display text-5xl text-ink tabular-nums">
              ₹{balance}
            </div>
          )}
        </div>

        <div className="grid grid-cols-2 gap-3">
          <Link
            to="/send"
            className="glass-panel rounded-xl p-4 text-center hover:bg-white/70 transition"
          >
            <div className="text-rose font-display text-lg mb-1">Send</div>
            <div className="text-xs text-muted">Transfer to a username</div>
          </Link>
          <Link
            to="/transactions"
            className="glass-panel rounded-xl p-4 text-center hover:bg-white/70 transition"
          >
            <div className="text-rose font-display text-lg mb-1">Activity</div>
            <div className="text-xs text-muted">Your transaction history</div>
          </Link>
        </div>
      </div>
    </div>
  );
}
