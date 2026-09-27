import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { InputBox } from '../components/InputBox';
import { UserAutocomplete } from '../components/UserAutocomplete';
import { Heading } from '../components/Heading';
import { SubHeading } from '../components/SubHeading';
import { GlassCard } from '../components/GlassCard';
import { ErrorBanner } from '../components/ErrorBanner';
import { useToast } from '../context/ToastContext';
import api from '../utils/api';

export function Split() {
  const [participants, setParticipants] = useState(['']);
  const [totalAmount, setTotalAmount] = useState('');
  const [note, setNote] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const navigate = useNavigate();
  const { showToast } = useToast();

  function updateParticipant(index, value) {
    setParticipants((prev) => prev.map((p, i) => (i === index ? value : p)));
  }

  function addParticipant() {
    setParticipants((prev) => [...prev, '']);
  }

  function removeParticipant(index) {
    setParticipants((prev) => prev.filter((_, i) => i !== index));
  }

  async function handleSplit() {
    setError('');
    setResult(null);
    const cleaned = participants.map((p) => p.trim()).filter(Boolean);

    if (cleaned.length === 0) {
      setError('Add at least one participant');
      return;
    }
    if (!totalAmount || Number(totalAmount) <= 0) {
      setError('Enter a valid total amount');
      return;
    }

    setLoading(true);
    try {
      const { data } = await api.post('/requests/split', {
        participants: cleaned,
        total_amount: totalAmount,
        note,
      });
      setResult(data.requests || []);
      showToast('Split requests sent');
      setParticipants(['']);
      setTotalAmount('');
      setNote('');
    } catch (err) {
      setError(err.response?.data?.error || 'Could not split the bill');
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
          <Heading label="Split a bill" />
          <SubHeading label="Divide a total evenly and request each person's share" />

          <ErrorBanner message={error} />

          {result && (
            <div className="bg-rose/10 border border-rose/30 rounded-xl p-4 mb-4 text-sm text-ink space-y-1">
              {result.map((r) => (
                <div key={r.id}>Requested ₹{r.amount} from {r.payer_username}</div>
              ))}
            </div>
          )}

          <div className="text-left space-y-1">
            <InputBox label="Total amount" placeholder="900" type="number" value={totalAmount} onChange={(e) => setTotalAmount(e.target.value)} />
            <InputBox label="What's it for (optional)" placeholder="Dinner" value={note} onChange={(e) => setNote(e.target.value)} />

            <div className="text-sm font-medium text-left py-2 text-ink/80">Split with</div>
            {participants.map((p, i) => (
              <div key={i} className="flex items-center gap-2 mb-2">
                <div className="flex-1">
                  <UserAutocomplete value={p} onChange={(v) => updateParticipant(i, v)} placeholder="Search a name or username" />
                </div>
                {participants.length > 1 && (
                  <button onClick={() => removeParticipant(i)} className="text-muted hover:text-rose transition text-sm mt-6" type="button">
                    ✕
                  </button>
                )}
              </div>
            ))}
            <button onClick={addParticipant} type="button" className="text-sm text-rose hover:underline">
              + Add another person
            </button>
          </div>

          <div className="pt-4">
            <Button label={loading ? 'Splitting…' : 'Split bill'} onClick={handleSplit} disabled={loading} />
          </div>
        </GlassCard>
      </div>
    </div>
  );
}
