import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { InputBox } from '../components/InputBox';
import { Heading } from '../components/Heading';
import { SubHeading } from '../components/SubHeading';
import { GlassCard } from '../components/GlassCard';
import { ErrorBanner } from '../components/ErrorBanner';
import api from '../utils/api';

export function Deposit() {
  const [amount, setAmount] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function handleDeposit() {
    setError('');
    setSuccess(null);

    if (!amount) {
      setError('Enter an amount');
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
        '/account/deposit',
        { amount },
        { headers: { 'Idempotency-Key': idempotencyKey } }
      );
      setSuccess(data);
      setAmount('');
    } catch (err) {
      setError(err.response?.data?.error || 'Deposit failed');
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
          <Heading label="Add money" />
          <SubHeading label="Top up your wallet balance" />

          <ErrorBanner message={error} />

          {success && (
            <div className="bg-rose/10 border border-rose/30 rounded-xl p-4 mb-4 text-sm text-ink">
              Added successfully. Transaction <span className="font-medium">{success.transaction_id.slice(0, 8)}</span>
            </div>
          )}

          <div className="text-left space-y-1">
            <InputBox label="Amount" placeholder="500" type="number" onChange={(e) => setAmount(e.target.value)} value={amount} />
          </div>

          <div className="pt-4">
            <Button label={loading ? 'Adding…' : 'Add money'} onClick={handleDeposit} disabled={loading} />
          </div>
        </GlassCard>
      </div>
    </div>
  );
}
