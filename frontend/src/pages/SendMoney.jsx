import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Button } from '../components/Button';
import { InputBox } from '../components/InputBox';
import { Heading } from '../components/Heading';
import { SubHeading } from '../components/SubHeading';
import { GlassCard } from '../components/GlassCard';
import { ErrorBanner } from '../components/ErrorBanner';
import { Avatar } from '../components/Avatar';
import api from '../utils/api';

export function SendMoney() {
  const [toUsername, setToUsername] = useState('');
  const [amount, setAmount] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

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
    try {
      const idempotencyKey = crypto.randomUUID();
      const { data } = await api.post(
        '/transfer/',
        { to_username: toUsername, amount },
        { headers: { 'Idempotency-Key': idempotencyKey } }
      );
      setSuccess(data);
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
        <button onClick={() => navigate('/dashboard')} className="text-sm text-muted hover:text-rose transition mb-4">
          ← Back
        </button>

        <GlassCard>
          <Heading label="Send money" />
          <SubHeading label="Transfer instantly to another PayWave user" />

          <ErrorBanner message={error} />

          {success && (
            <div className="bg-rose/10 border border-rose/30 rounded-xl p-4 mb-4 flex items-center gap-3">
              <Avatar name={toUsername} size={36} />
              <div className="text-sm text-ink">
                Sent successfully. Transaction <span className="font-medium">{success.transaction_id.slice(0, 8)}</span>
              </div>
            </div>
          )}

          <div className="text-left space-y-1">
            <InputBox label="To username" placeholder="priya" onChange={(e) => setToUsername(e.target.value)} value={toUsername} />
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
