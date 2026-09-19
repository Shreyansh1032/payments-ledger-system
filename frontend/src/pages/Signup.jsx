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

export function Signup() {
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function handleSignup() {
    setError('');

    if (!firstName || !lastName || !username || !password) {
      setError('All fields are required');
      return;
    }
    if (password.length < 6) {
      setError('Password must be at least 6 characters');
      return;
    }

    setLoading(true);
    try {
      const { data } = await api.post('/auth/signup', {
        username,
        first_name: firstName,
        last_name: lastName,
        password,
      });
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('username', username);
      localStorage.setItem('first_name', firstName);
      navigate('/dashboard');
    } catch (err) {
      setError(err.response?.data?.error || 'Something went wrong. Try again.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="aura-backdrop min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-sm">
        <GlassCard className="text-center">
          <Heading label="Create your account" />
          <SubHeading label="Set up your wallet in a minute" />

          <ErrorBanner message={error} />

          <div className="text-left space-y-1">
            <InputBox label="First name" placeholder="Shreyansh" onChange={(e) => setFirstName(e.target.value)} />
            <InputBox label="Last name" placeholder="Sinha" onChange={(e) => setLastName(e.target.value)} />
            <InputBox label="Username" placeholder="shreyansh" onChange={(e) => setUsername(e.target.value)} />
            <InputBox label="Password" placeholder="At least 6 characters" type="password" onChange={(e) => setPassword(e.target.value)} />
          </div>

          <div className="pt-4">
            <Button label={loading ? 'Creating account…' : 'Sign up'} onClick={handleSignup} disabled={loading} />
          </div>

          <BottomWarning label="Already have an account?" buttonText="Sign in" to="/signin" />
        </GlassCard>
      </div>
    </div>
  );
}
