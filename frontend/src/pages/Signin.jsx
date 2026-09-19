import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { InputBox } from '../components/InputBox';
import { Heading } from '../components/Heading';
import { SubHeading } from '../components/SubHeading';
import { BottomWarning } from '../components/BottomWarning';
import { GlassCard } from '../components/GlassCard';
import { ErrorBanner } from '../components/ErrorBanner';
import api from '../utils/api';

export function Signin() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function handleSignin() {
    setError('');

    if (!username || !password) {
      setError('Enter your username and password');
      return;
    }

    setLoading(true);
    try {
      const { data } = await api.post('/auth/login', { username, password });
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('username', username);
      navigate('/dashboard');
    } catch (err) {
      setError(err.response?.data?.error || 'Invalid username or password');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="aura-backdrop min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-sm">
        <GlassCard className="text-center">
          <Heading label="Welcome back" />
          <SubHeading label="Sign in to continue to your wallet" />

          <ErrorBanner message={error} />

          <div className="text-left space-y-1">
            <InputBox label="Username" placeholder="shreyansh" onChange={(e) => setUsername(e.target.value)} />
            <InputBox label="Password" placeholder="Your password" type="password" onChange={(e) => setPassword(e.target.value)} />
          </div>

          <div className="pt-4">
            <Button label={loading ? 'Signing in…' : 'Sign in'} onClick={handleSignin} disabled={loading} />
          </div>

          <BottomWarning label="Don't have an account?" buttonText="Sign up" to="/signup" />
        </GlassCard>
      </div>
    </div>
  );
}
