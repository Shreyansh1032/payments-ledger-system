import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import QRCode from 'qrcode';
import { Avatar } from '../components/Avatar';
import { GlassCard } from '../components/GlassCard';
import { getPayWaveId } from '../utils/paywaveId';

export function Profile() {
  const [qrDataUrl, setQrDataUrl] = useState('');
  const [copied, setCopied] = useState(false);
  const navigate = useNavigate();

  const username = localStorage.getItem('username') || '';
  const firstName = localStorage.getItem('first_name') || username;
  const payWaveId = getPayWaveId(username);

  useEffect(() => {
    const payLink = `${window.location.origin}/send?to=${encodeURIComponent(username)}`;
    QRCode.toDataURL(payLink, {
      width: 240,
      margin: 1,
      color: { dark: '#241521', light: '#00000000' },
    }).then(setQrDataUrl);
  }, [username]);

  function handleCopy() {
    navigator.clipboard.writeText(payWaveId);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  return (
    <div className="aura-backdrop min-h-screen px-4 py-8">
      <div className="max-w-sm mx-auto">
        <button onClick={() => navigate('/dashboard')} className="text-sm text-muted hover:text-rose transition mb-4">
          ← Back
        </button>

        <GlassCard className="text-center">
          <div className="flex justify-center mb-3">
            <Avatar name={firstName} size={56} />
          </div>
          <div className="font-display text-xl text-ink">{firstName}</div>
          <div className="text-sm text-muted mb-4">{payWaveId}</div>

          {qrDataUrl && (
            <img src={qrDataUrl} alt="Your PayWave QR code" className="mx-auto rounded-xl" />
          )}

          <div className="text-xs text-muted mt-3 mb-4">
            Others can scan this to send you money directly
          </div>

          <button
            onClick={handleCopy}
            className="w-full text-rose bg-transparent border border-rose/30 hover:bg-blush font-sans font-semibold rounded-xl text-sm px-5 py-2.5 transition-colors"
          >
            {copied ? 'Copied!' : 'Copy PayWave ID'}
          </button>
        </GlassCard>
      </div>
    </div>
  );
}
