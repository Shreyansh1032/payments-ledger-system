import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../utils/api';

const GOOGLE_CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID;

export function GoogleSignInButton({ onError }) {
  const buttonRef = useRef(null);
  const submitRef = useRef(null);
  const navigate = useNavigate();

  const [pendingCredential, setPendingCredential] = useState(null);
  const [code, setCode] = useState('');
  const [codeError, setCodeError] = useState('');
  const [loading, setLoading] = useState(false);

  async function submit(credential, totpCode) {
    const payload = { id_token: credential };
    if (totpCode) payload.totp_code = totpCode;

    setLoading(true);
    try {
      const { data } = await api.post('/auth/google', payload);

      if (data.totp_required) {
        setPendingCredential(credential);
        setCodeError('');
        return;
      }

      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('username', data.username);
      localStorage.setItem('first_name', data.first_name);
      navigate('/dashboard');
    } catch (err) {
      const message = err.response?.data?.error || 'Google sign-in failed';
      if (totpCode) {
        setCodeError(message);
        setCode('');
      } else {
        onError?.(message);
      }
    } finally {
      setLoading(false);
    }
  }
  submitRef.current = submit;

  useEffect(() => {
    if (!GOOGLE_CLIENT_ID) return;

    function handleCredentialResponse(response) {
      submitRef.current(response.credential);
    }

    function initialize() {
      if (!window.google || !buttonRef.current) return;
      window.google.accounts.id.initialize({
        client_id: GOOGLE_CLIENT_ID,
        callback: handleCredentialResponse,
      });
      window.google.accounts.id.renderButton(buttonRef.current, {
        theme: 'outline',
        size: 'large',
        width: 320,
        shape: 'pill',
      });
    }

    if (window.google) {
      initialize();
      return;
    }

    const existingScript = document.getElementById('google-identity-script');
    if (existingScript) {
      existingScript.addEventListener('load', initialize);
      return () => existingScript.removeEventListener('load', initialize);
    }

    const script = document.createElement('script');
    script.id = 'google-identity-script';
    script.src = 'https://accounts.google.com/gsi/client';
    script.async = true;
    script.defer = true;
    script.onload = initialize;
    document.head.appendChild(script);
  }, []);

  if (!GOOGLE_CLIENT_ID) return null;

  function handleVerify() {
    if (!code) {
      setCodeError('Enter the 6-digit code from your authenticator app');
      return;
    }
    submit(pendingCredential, code);
  }

  function handleCancel() {
    setPendingCredential(null);
    setCode('');
    setCodeError('');
  }

  return (
    <>
      <div ref={buttonRef} className={pendingCredential ? 'hidden' : 'flex justify-center my-4'} />

      {pendingCredential && (
        <div className="my-4 text-left space-y-3">
          <p className="text-sm text-center opacity-80">
            Two-factor authentication is on for this account. Enter the code from your authenticator app.
          </p>
          {codeError && <p className="text-sm text-center text-red-500">{codeError}</p>}
          <input
            className="w-full rounded-xl border border-white/60 bg-white/50 px-4 py-3 text-center tracking-widest outline-none focus:ring-2 focus:ring-pink-300"
            placeholder="123456"
            inputMode="numeric"
            maxLength={6}
            value={code}
            onChange={(e) => setCode(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleVerify()}
          />
          <button
            type="button"
            onClick={handleVerify}
            disabled={loading}
            className="w-full rounded-xl bg-pink-500 py-3 font-medium text-white disabled:opacity-60"
          >
            {loading ? 'Verifying…' : 'Verify'}
          </button>
          <button type="button" onClick={handleCancel} className="w-full text-sm opacity-70">
            Cancel
          </button>
        </div>
      )}
    </>
  );
}