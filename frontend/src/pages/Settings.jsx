import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { InputBox } from '../components/InputBox';
import { Avatar } from '../components/Avatar';
import { GlassCard } from '../components/GlassCard';
import { ErrorBanner } from '../components/ErrorBanner';
import { getPayWaveId } from '../utils/paywaveId';
import { useToast } from '../context/ToastContext';
import api from '../utils/api';

export function Settings() {
  const [profile, setProfile] = useState(null);
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { showToast } = useToast();

  useEffect(() => {
    api.get('/account/profile').then(({ data }) => setProfile(data)).catch(() => {});
  }, []);

  async function handleChangePassword() {
    setError('');
    if (!currentPassword || !newPassword) {
      setError('Fill in both password fields');
      return;
    }
    if (newPassword.length < 6) {
      setError('New password must be at least 6 characters');
      return;
    }
    if (newPassword !== confirmPassword) {
      setError('New passwords do not match');
      return;
    }

    setLoading(true);
    try {
      await api.put('/account/password', { current_password: currentPassword, new_password: newPassword });
      showToast('Password updated');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err) {
      setError(err.response?.data?.error || 'Could not change password');
    } finally {
      setLoading(false);
    }
  }

  const memberSince = profile?.created_at
    ? new Date(profile.created_at).toLocaleDateString('en-IN', { month: 'long', year: 'numeric' })
    : '';

  return (
    <div className="aura-backdrop min-h-screen px-4 py-8">
      <div className="max-w-sm mx-auto">
        <button onClick={() => navigate('/dashboard')} className="text-sm text-muted hover:text-rose transition mb-4">
          ← Back
        </button>

        <div className="font-display text-2xl text-ink mb-6">Settings</div>

        {profile && (
          <GlassCard className="mb-6 text-center">
            <div className="flex justify-center mb-3">
              <Avatar name={profile.first_name} size={48} />
            </div>
            <div className="font-display text-lg text-ink">{profile.first_name} {profile.last_name}</div>
            <div className="text-sm text-muted">{getPayWaveId(profile.username)}</div>
            {memberSince && <div className="text-xs text-muted mt-2">Member since {memberSince}</div>}
          </GlassCard>
        )}

        <GlassCard>
          <div className="font-display text-lg text-ink mb-3 text-left">Change password</div>

          <ErrorBanner message={error} />

          <div className="text-left space-y-1">
            <InputBox label="Current password" type="password" value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)} placeholder="Current password" />
            <InputBox label="New password" type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} placeholder="At least 6 characters" />
            <InputBox label="Confirm new password" type="password" value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} placeholder="Repeat new password" />
          </div>

          <div className="pt-4">
            <Button label={loading ? 'Updating…' : 'Update password'} onClick={handleChangePassword} disabled={loading} />
          </div>
        </GlassCard>
      </div>
    </div>
  );
}
