import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Button } from '../components/Button';
import { InputBox } from '../components/InputBox';
import { UserAutocomplete } from '../components/UserAutocomplete';
import { Heading } from '../components/Heading';
import { SubHeading } from '../components/SubHeading';
import { GlassCard } from '../components/GlassCard';
import { ErrorBanner } from '../components/ErrorBanner';
import { Avatar } from '../components/Avatar';
import api from '../utils/api';

export function SendMoney() {
  const [searchParams] = useSearchParams();
  const [toUsername, setToUsername] = useState('');
  const [amount, setAmount] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(null);
  const [loading, setLoading] = useState(false);
  const [favorites, setFavorites] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    const prefill = searchParams.get('to');
    if (prefill) setToUsername(prefill);
  }, [searchParams]);

  useEffect(() => {
    api.get('/favorites').then(({ data }) => setFavorites(data.favorites || [])).catch(() => {});
  }, []);

  async function handleSend() {
    setError('');
    setSuccess(null);

    if (!toUsername || !amount) {
      setError('Enter a username and amount');
      return;
    }
    if (Number(amount) <= 0) {
      setError('Amount must be positive');
      return;
    }

    setLoading(true);
    const recipient = toUsername;
    try {
      const idempotencyKey = crypto.randomUUID();
      const { data } = await api.post(
        '/transfer/',
        { to_username: recipient, amount },
        { headers: { 'Idempotency-Key': idempotencyKey } }
      );
      setSuccess({ ...data, recipient });
      setToUsername('');
      setAmount('');
    } catch (err) {
      setError(err.response?.data?.error || 'Transfer failed');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="aura-backdrop min-h-screen px-4 py-8">
      <div className="max-w-sm mx-auto">
        <div className="flex items-center justify-between mb-4">
          <button onClick={() => navigate('/dashboard')} className="text-sm text-muted hover:text-rose transition">
            ← Back
          </button>
          <button onClick={() => navigate('/scan')} className="text-sm text-rose font-medium hover:underline">
            Scan QR
          </button>
        </div>

        {favorites.length > 0 && (
          <div className="flex gap-3 overflow-x-auto pb-2 mb-4">
            {favorites.map((f) => (
              <button
                key={f.username}
                onClick={() => setToUsername(f.username)}
                className="flex flex-col items-center gap-1 shrink-0"
              >
                <Avatar name={f.first_name} size={44} />
                <div className="text-xs text-muted truncate max-w-[56px]">{f.first_name}</div>
              </button>
            ))}
          </div>
        )}

        <GlassCard>
          <Heading label="Send money" />
          <SubHeading label="Transfer instantly to another PayWave user" />

          <ErrorBanner message={error} />

          {success && (
            <div className="bg-rose/10 border border-rose/30 rounded-xl p-4 mb-4 flex items-center gap-3">
              <Avatar name={success.recipient} size={36} />
              <div className="text-sm text-ink">
                Sent to {success.recipient}. Transaction <span className="font-medium">{success.transaction_id.slice(0, 8)}</span>
              </div>
            </div>
          )}

          <div className="text-left space-y-1">
            <UserAutocomplete label="To" value={toUsername} onChange={setToUsername} placeholder="Search a name or username" />
            <InputBox label="Amount" placeholder="100" type="number" onChange={(e) => setAmount(e.target.value)} value={amount} />
          </div>

          <div className="pt-4">
            <Button label={loading ? 'Sending…' : 'Send'} onClick={handleSend} disabled={loading} />
          </div>
        </GlassCard>
      </div>
    </div>
  );
}
