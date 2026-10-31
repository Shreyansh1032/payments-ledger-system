import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import QRCode from 'qrcode';
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

  const [totpSetup, setTotpSetup] = useState(null);
  const [totpQrDataUrl, setTotpQrDataUrl] = useState('');
  const [totpConfirmCode, setTotpConfirmCode] = useState('');
  const [totpDisableCode, setTotpDisableCode] = useState('');
  const [totpError, setTotpError] = useState('');
  const [totpLoading, setTotpLoading] = useState(false);

  function loadProfile() {
    api.get('/account/profile').then(({ data }) => setProfile(data)).catch(() => {});
  }

  useEffect(() => {
    loadProfile();
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

  async function handleStartTotpSetup() {
    setTotpError('');
    setTotpLoading(true);
    try {
      const { data } = await api.post('/account/2fa/setup');
      setTotpSetup(data);
      const dataUrl = await QRCode.toDataURL(data.otpauth_url, { width: 220, margin: 1 });
      setTotpQrDataUrl(dataUrl);
    } catch (err) {
      setTotpError(err.response?.data?.error || 'Could not start 2FA setup');
    } finally {
      setTotpLoading(false);
    }
  }

  async function handleConfirmTotp() {
    setTotpError('');
    if (!totpConfirmCode) {
      setTotpError('Enter the 6-digit code from your app');
      return;
    }
    setTotpLoading(true);
    try {
      await api.post('/account/2fa/confirm', { code: totpConfirmCode });
      showToast('Two-factor authentication enabled');
      setTotpSetup(null);
      setTotpQrDataUrl('');
      setTotpConfirmCode('');
      loadProfile();
    } catch (err) {
      setTotpError(err.response?.data?.error || 'Invalid code');
    } finally {
      setTotpLoading(false);
    }
  }

  async function handleDisableTotp() {
    setTotpError('');
    if (!totpDisableCode) {
      setTotpError('Enter your current 6-digit code to disable');
      return;
    }
    setTotpLoading(true);
    try {
      await api.post('/account/2fa/disable', { code: totpDisableCode });
      showToast('Two-factor authentication disabled');
      setTotpDisableCode('');
      loadProfile();
    } catch (err) {
      setTotpError(err.response?.data?.error || 'Invalid code');
    } finally {
      setTotpLoading(false);
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

        <GlassCard className="mb-6">
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

        {profile && (
          <GlassCard>
            <div className="font-display text-lg text-ink mb-1 text-left">Two-factor authentication</div>
            <div className="text-sm text-muted text-left mb-3">
              {profile.totp_enabled ? 'Enabled — an authenticator code is required at sign-in' : 'Add an extra layer of security to your account'}
            </div>

            <ErrorBanner message={totpError} />

            {profile.totp_enabled ? (
              <div className="text-left space-y-1">
                <InputBox label="Enter current code to disable" placeholder="123456" value={totpDisableCode} onChange={(e) => setTotpDisableCode(e.target.value)} />
                <div className="pt-4">
                  <Button label={totpLoading ? 'Disabling…' : 'Disable 2FA'} onClick={handleDisableTotp} disabled={totpLoading} variant="ghost" />
                </div>
              </div>
            ) : !totpSetup ? (
              <Button label={totpLoading ? 'Starting…' : 'Set up 2FA'} onClick={handleStartTotpSetup} disabled={totpLoading} />
            ) : (
              <div className="text-left">
                <div className="text-sm text-muted mb-3">Scan this with Google Authenticator, Authy, or any TOTP app</div>
                {totpQrDataUrl && <img src={totpQrDataUrl} alt="2FA setup QR code" className="mx-auto rounded-xl mb-3" />}
                <div className="text-xs text-muted text-center mb-4 break-all">Or enter manually: {totpSetup.secret}</div>
                <InputBox label="Enter the 6-digit code to confirm" placeholder="123456" value={totpConfirmCode} onChange={(e) => setTotpConfirmCode(e.target.value)} />
                <div className="pt-4">
                  <Button label={totpLoading ? 'Confirming…' : 'Confirm and enable'} onClick={handleConfirmTotp} disabled={totpLoading} />
                </div>
              </div>
            )}
          </GlassCard>
        )}
      </div>
    </div>
  );
}
