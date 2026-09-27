import { useEffect, useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Avatar } from '../components/Avatar';
import { ErrorBanner } from '../components/ErrorBanner';
import { ThemeToggle } from '../components/ThemeToggle';
import { formatDate } from '../utils/date';
import api from '../utils/api';

export function Dashboard() {
  const [balance, setBalance] = useState(null);
  const [recent, setRecent] = useState([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  const username = localStorage.getItem('username') || '';
  const firstName = localStorage.getItem('first_name') || username;

  useEffect(() => {
    async function fetchData() {
      try {
        const [balanceRes, txnRes] = await Promise.all([
          api.get('/account/balance'),
          api.get('/transactions/'),
        ]);
        setBalance(balanceRes.data.balance);
        setRecent((txnRes.data.transactions || []).slice(0, 4));
      } catch (err) {
        setError(err.response?.data?.error || 'Could not load your account');
      } finally {
        setLoading(false);
      }
    }
    fetchData();
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
    <div className="aura-backdrop min-h-screen">
      <header className="max-w-5xl mx-auto px-4 sm:px-6 pt-6 pb-4 flex items-center justify-between flex-wrap gap-3">
        <div className="font-display text-xl text-rose">PayWave</div>
        <div className="flex items-center gap-4">
          <Link to="/profile" className="flex items-center gap-2 min-w-0">
            <Avatar name={firstName} size={36} />
            <div className="text-sm text-ink font-medium hidden sm:block truncate">{firstName}</div>
          </Link>
          <ThemeToggle />
          <button onClick={handleLogout} className="text-sm text-muted hover:text-rose transition shrink-0">
            Log out
          </button>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 sm:px-6 pb-12">
        <ErrorBanner message={error} />

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
          <div className="max-w-md w-full mx-auto lg:mx-0">
            <div className="glass-panel rounded-2xl p-6 sm:p-8 text-center mb-6">
              <div className="text-sm text-muted mb-2">Your balance</div>
              {loading ? (
                <div className="font-display text-4xl text-ink/40">…</div>
              ) : (
                <div className="font-display text-4xl sm:text-5xl text-ink tabular-nums break-all">
                  ₹{balance}
                </div>
              )}
            </div>

            <div className="grid grid-cols-2 gap-3 mb-6">
              <Link to="/deposit" className="glass-panel rounded-xl p-4 text-center hover:bg-white/70 dark:hover:bg-white/5 transition">
                <div className="text-rose font-display text-lg mb-1">Add money</div>
                <div className="text-xs text-muted">Top up balance</div>
              </Link>
              <Link to="/send" className="glass-panel rounded-xl p-4 text-center hover:bg-white/70 dark:hover:bg-white/5 transition">
                <div className="text-rose font-display text-lg mb-1">Send</div>
                <div className="text-xs text-muted">To a username</div>
              </Link>
              <Link to="/profile" className="glass-panel rounded-xl p-4 text-center hover:bg-white/70 dark:hover:bg-white/5 transition">
                <div className="text-rose font-display text-lg mb-1">My QR</div>
                <div className="text-xs text-muted">Get paid instantly</div>
              </Link>
              <Link to="/transactions" className="glass-panel rounded-xl p-4 text-center hover:bg-white/70 dark:hover:bg-white/5 transition">
                <div className="text-rose font-display text-lg mb-1">Activity</div>
                <div className="text-xs text-muted">History</div>
              </Link>
            </div>

            <div className="flex flex-wrap gap-x-4 gap-y-2 justify-center text-sm">
              <Link to="/requests" className="text-muted hover:text-rose transition">Requests</Link>
              <Link to="/split" className="text-muted hover:text-rose transition">Split a bill</Link>
              <Link to="/insights" className="text-muted hover:text-rose transition">Insights</Link>
              <Link to="/settings" className="text-muted hover:text-rose transition">Settings</Link>
            </div>
          </div>

          <div className="max-w-md w-full mx-auto lg:mx-0">
            <div className="flex items-center justify-between mb-3">
              <div className="font-display text-lg text-ink">Recent activity</div>
              <Link to="/transactions" className="text-sm text-rose hover:underline">View all</Link>
            </div>

            {loading && (
              <div className="text-center text-muted text-sm py-8">Loading…</div>
            )}

            {!loading && recent.length === 0 && (
              <div className="glass-panel rounded-2xl p-6 text-center">
                <div className="text-sm text-muted">No activity yet — send or add money to get started</div>
              </div>
            )}

            {!loading && recent.length > 0 && (
              <div className="space-y-2">
                {recent.map((t) => (
                  <div key={t.id} className="glass-panel rounded-xl p-3 flex items-center gap-3">
                    <Avatar name={t.counterparty} size={36} />
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-medium text-ink truncate">
                        {t.direction === 'sent' ? `To ${t.counterparty}` : `From ${t.counterparty}`}
                      </div>
                      <div className="text-xs text-muted">{formatDate(t.created_at)}</div>
                    </div>
                    <div className={`font-display text-base tabular-nums ${t.direction === 'sent' ? 'text-ink' : 'text-rose'}`}>
                      {t.direction === 'sent' ? '-' : '+'}₹{t.amount}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
