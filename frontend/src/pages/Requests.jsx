import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { InputBox } from '../components/InputBox';
import { UserAutocomplete } from '../components/UserAutocomplete';
import { GlassCard } from '../components/GlassCard';
import { ErrorBanner } from '../components/ErrorBanner';
import { Avatar } from '../components/Avatar';
import { formatDate } from '../utils/date';
import { useToast } from '../context/ToastContext';
import api from '../utils/api';

export function Requests() {
  const [incoming, setIncoming] = useState([]);
  const [outgoing, setOutgoing] = useState([]);
  const [payerUsername, setPayerUsername] = useState('');
  const [amount, setAmount] = useState('');
  const [note, setNote] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { showToast } = useToast();

  async function loadRequests() {
    try {
      const [inc, out] = await Promise.all([
        api.get('/requests/incoming'),
        api.get('/requests/outgoing'),
      ]);
      setIncoming(inc.data.requests || []);
      setOutgoing(out.data.requests || []);
    } catch (err) {
      setError(err.response?.data?.error || 'Could not load requests');
    }
  }

  useEffect(() => {
    loadRequests();
  }, []);

  async function handleCreate() {
    setError('');
    if (!payerUsername || !amount) {
      setError('Enter a username and amount');
      return;
    }
    setLoading(true);
    try {
      await api.post('/requests/', { payer_username: payerUsername, amount, note });
      showToast(`Requested ₹${amount} from ${payerUsername}`);
      setPayerUsername('');
      setAmount('');
      setNote('');
      loadRequests();
    } catch (err) {
      setError(err.response?.data?.error || 'Could not create request');
    } finally {
      setLoading(false);
    }
  }

  async function handleApprove(id) {
    try {
      await api.post(`/requests/${id}/approve`);
      showToast('Payment sent');
      loadRequests();
    } catch (err) {
      showToast(err.response?.data?.error || 'Could not approve request', 'error');
    }
  }

  async function handleDecline(id) {
    try {
      await api.post(`/requests/${id}/decline`);
      showToast('Request declined');
      loadRequests();
    } catch (err) {
      showToast(err.response?.data?.error || 'Could not decline request', 'error');
    }
  }

  return (
    <div className="aura-backdrop min-h-screen px-4 py-8">
      <div className="max-w-md mx-auto">
        <button onClick={() => navigate('/dashboard')} className="text-sm text-muted hover:text-rose transition mb-4">
          ← Back
        </button>

        <div className="font-display text-2xl text-ink mb-6">Requests</div>

        <ErrorBanner message={error} />

        <GlassCard className="mb-6">
          <div className="font-display text-lg text-ink mb-3">Request money</div>
          <div className="text-left space-y-1">
            <UserAutocomplete label="From" value={payerUsername} onChange={setPayerUsername} placeholder="Search a name or username" />
            <InputBox label="Amount" placeholder="100" type="number" value={amount} onChange={(e) => setAmount(e.target.value)} />
            <InputBox label="Note (optional)" placeholder="Dinner split" value={note} onChange={(e) => setNote(e.target.value)} />
          </div>
          <div className="pt-4">
            <Button label={loading ? 'Requesting…' : 'Send request'} onClick={handleCreate} disabled={loading} />
          </div>
        </GlassCard>

        {incoming.length > 0 && (
          <div className="mb-6">
            <div className="font-display text-lg text-ink mb-3">Requests for you</div>
            <div className="space-y-2">
              {incoming.map((r) => (
                <div key={r.id} className="glass-panel rounded-xl p-4 flex items-center gap-3">
                  <Avatar name={r.requester_username} size={36} />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium text-ink truncate">{r.requester_username} wants ₹{r.amount}</div>
                    {r.note && <div className="text-xs text-muted truncate">{r.note}</div>}
                    <div className="text-xs text-muted">{formatDate(r.created_at)}</div>
                  </div>
                  {r.status === 'pending' ? (
                    <div className="flex gap-2 shrink-0">
                      <button onClick={() => handleApprove(r.id)} className="text-xs bg-rose text-white rounded-lg px-3 py-1.5 hover:bg-rose-dark transition">Pay</button>
                      <button onClick={() => handleDecline(r.id)} className="text-xs border border-rose/30 text-rose rounded-lg px-3 py-1.5 hover:bg-blush transition">Decline</button>
                    </div>
                  ) : (
                    <div className="text-xs text-muted shrink-0 capitalize">{r.status}</div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        <div>
          <div className="font-display text-lg text-ink mb-3">Your requests</div>
          {outgoing.length === 0 ? (
            <div className="glass-panel rounded-2xl p-6 text-center text-sm text-muted">
              You haven't requested any money yet
            </div>
          ) : (
            <div className="space-y-2">
              {outgoing.map((r) => (
                <div key={r.id} className="glass-panel rounded-xl p-4 flex items-center gap-3">
                  <Avatar name={r.payer_username} size={36} />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium text-ink truncate">₹{r.amount} from {r.payer_username}</div>
                    <div className="text-xs text-muted">{formatDate(r.created_at)}</div>
                  </div>
                  <div className="text-xs text-muted capitalize shrink-0">{r.status}</div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
