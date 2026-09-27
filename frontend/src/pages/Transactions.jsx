import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Avatar } from '../components/Avatar';
import { ErrorBanner } from '../components/ErrorBanner';
import { formatDate } from '../utils/date';
import api from '../utils/api';

export function Transactions() {
  const [transactions, setTransactions] = useState([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    async function fetchHistory() {
      try {
        const { data } = await api.get('/transactions/');
        setTransactions(data.transactions || []);
      } catch (err) {
        setError(err.response?.data?.error || 'Could not load transactions');
      } finally {
        setLoading(false);
      }
    }
    fetchHistory();
  }, []);

  return (
    <div className="aura-backdrop min-h-screen px-4 py-8">
      <div className="max-w-md mx-auto">
        <button onClick={() => navigate('/dashboard')} className="text-sm text-muted hover:text-rose transition mb-4">
          ← Back
        </button>

        <div className="font-display text-2xl text-ink mb-1">Activity</div>
        <div className="text-sm text-muted mb-6">Your recent transfers</div>

        <ErrorBanner message={error} />

        {loading && (
          <div className="text-center text-muted text-sm py-12">Loading…</div>
        )}

        {!loading && transactions.length === 0 && !error && (
          <div className="glass-panel rounded-2xl p-8 text-center">
            <div className="font-display text-lg text-ink mb-1">No transactions yet</div>
            <div className="text-sm text-muted">Send money to someone and it'll show up here</div>
          </div>
        )}

        {!loading && transactions.length > 0 && (
          <div className="space-y-2">
            {transactions.map((t) => (
              <div key={t.id} className="glass-panel rounded-xl p-4 flex items-center gap-3">
                <Avatar name={t.counterparty} size={40} />
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium text-ink truncate">
                    {t.direction === 'sent' ? `To ${t.counterparty}` : `From ${t.counterparty}`}
                  </div>
                  <div className="text-xs text-muted">{formatDate(t.created_at)}</div>
                </div>
                <div
                  className={`font-display text-lg tabular-nums ${
                    t.direction === 'sent' ? 'text-ink' : 'text-rose'
                  }`}
                >
                  {t.direction === 'sent' ? '-' : '+'}₹{t.amount}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
