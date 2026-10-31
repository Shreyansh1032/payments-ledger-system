import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { InputBox } from '../components/InputBox';
import { Heading } from '../components/Heading';
import { SubHeading } from '../components/SubHeading';
import { BottomWarning } from '../components/BottomWarning';
import { GlassCard } from '../components/GlassCard';
import { ErrorBanner } from '../components/ErrorBanner';
import { GoogleSignInButton } from '../components/GoogleSignInButton';
import api from '../utils/api';

export function Signin() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [totpCode, setTotpCode] = useState('');
  const [needsTotp, setNeedsTotp] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const existingToken = localStorage.getItem('access_token');
    if (existingToken) {
      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken) {
        api.post('/auth/logout', { refresh_token: refreshToken }).catch(() => {});
      }
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('username');
      localStorage.removeItem('first_name');
    }
  }, []);

  async function handleSignin() {
    setError('');

    if (!username || !password) {
      setError('Enter your username and password');
      return;
    }
    if (needsTotp && !totpCode) {
      setError('Enter the 6-digit code from your authenticator app');
      return;
    }

    setLoading(true);
    try {
      const payload = { username, password };
      if (needsTotp) payload.totp_code = totpCode;

      const { data } = await api.post('/auth/login', payload);

      if (data.totp_required) {
        setNeedsTotp(true);
        setLoading(false);
        return;
      }

      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('username', username);
      navigate('/dashboard');
    } catch (err) {
      setError(err.response?.data?.error || 'Invalid username or password');
      if (needsTotp) setTotpCode('');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="aura-backdrop min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-sm">
        <GlassCard className="text-center">
          <Heading label={needsTotp ? 'Enter your code' : 'Welcome back'} />
          <SubHeading label={needsTotp ? 'Open your authenticator app' : 'Sign in to continue to your wallet'} />

          <ErrorBanner message={error} />

          {!needsTotp ? (
            <div className="text-left space-y-1">
              <InputBox label="Username" placeholder="shreyansh" value={username} onChange={(e) => setUsername(e.target.value)} />
              <InputBox label="Password" placeholder="Your password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
            </div>
          ) : (
            <div className="text-left space-y-1">
              <InputBox label="6-digit code" placeholder="123456" value={totpCode} onChange={(e) => setTotpCode(e.target.value)} />
            </div>
          )}

          <div className="pt-4">
            <Button label={loading ? 'Signing in…' : needsTotp ? 'Verify' : 'Sign in'} onClick={handleSignin} disabled={loading} />
          </div>

          {needsTotp && (
            <button
              onClick={() => { setNeedsTotp(false); setTotpCode(''); setError(''); }}
              className="text-sm text-muted hover:text-rose transition mt-2"
            >
              ← Back
            </button>
          )}

          {!needsTotp && (
            <>
              <div className="flex items-center gap-3 my-4">
                <div className="h-px bg-ink/10 flex-1" />
                <div className="text-xs text-muted">or</div>
                <div className="h-px bg-ink/10 flex-1" />
              </div>
              <GoogleSignInButton onError={setError} />
            </>
          )}

          <BottomWarning label="Don't have an account?" buttonText="Sign up" to="/signup" />
        </GlassCard>
      </div>
    </div>
  );
}
