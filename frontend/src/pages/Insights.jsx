import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Avatar } from '../components/Avatar';
import { ErrorBanner } from '../components/ErrorBanner';
import api from '../utils/api';

export function Insights() {
  const [insights, setInsights] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    api.get('/insights/monthly')
      .then(({ data }) => setInsights(data))
      .catch((err) => setError(err.response?.data?.error || 'Could not load insights'))
      .finally(() => setLoading(false));
  }, []);

  const monthLabel = new Date().toLocaleDateString('en-IN', { month: 'long' });

  return (
    <div className="aura-backdrop min-h-screen px-4 py-8">
      <div className="max-w-md mx-auto">
        <button onClick={() => navigate('/dashboard')} className="text-sm text-muted hover:text-rose transition mb-4">
          ← Back
        </button>

        <div className="font-display text-2xl text-ink mb-1">Insights</div>
        <div className="text-sm text-muted mb-6">Your activity in {monthLabel}</div>

        <ErrorBanner message={error} />

        {loading && <div className="text-center text-muted text-sm py-12">Loading…</div>}

        {!loading && insights && (
          <>
            <div className="grid grid-cols-2 gap-3 mb-6">
              <div className="glass-panel rounded-2xl p-5 text-center">
                <div className="text-xs text-muted mb-1">Sent</div>
                <div className="font-display text-2xl text-ink tabular-nums">₹{insights.total_sent}</div>
              </div>
              <div className="glass-panel rounded-2xl p-5 text-center">
                <div className="text-xs text-muted mb-1">Received</div>
                <div className="font-display text-2xl text-rose tabular-nums">₹{insights.total_received}</div>
              </div>
            </div>

            <div className="font-display text-lg text-ink mb-3">Top recipients</div>
            {insights.top_recipients.length === 0 ? (
              <div className="glass-panel rounded-2xl p-6 text-center text-sm text-muted">
                No transfers sent this month yet
              </div>
            ) : (
              <div className="space-y-2">
                {insights.top_recipients.map((r) => (
                  <div key={r.username} className="glass-panel rounded-xl p-3 flex items-center gap-3">
                    <Avatar name={r.username} size={32} />
                    <div className="flex-1 text-sm text-ink">{r.username}</div>
                    <div className="font-display text-sm tabular-nums text-ink">₹{r.total}</div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
