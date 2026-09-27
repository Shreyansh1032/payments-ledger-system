import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Html5Qrcode } from 'html5-qrcode';
import { ErrorBanner } from '../components/ErrorBanner';

export function Scan() {
  const navigate = useNavigate();
  const [error, setError] = useState('');
  const scannerRef = useRef(null);
  const containerId = 'qr-scan-region';

  useEffect(() => {
    const scanner = new Html5Qrcode(containerId);
    scannerRef.current = scanner;

    scanner
      .start(
        { facingMode: 'environment' },
        { fps: 10, qrbox: 220 },
        (decodedText) => {
          scanner.stop().then(() => {
            try {
              const url = new URL(decodedText);
              const to = url.searchParams.get('to');
              if (to) {
                navigate(`/send?to=${encodeURIComponent(to)}`);
                return;
              }
            } catch {
              // not a URL -- fall through and treat it as a raw username
            }
            navigate(`/send?to=${encodeURIComponent(decodedText)}`);
          });
        },
        () => {
          // per-frame scan miss, expected while the camera searches
        }
      )
      .catch(() => {
        setError('Could not access the camera. Check your browser permissions.');
      });

    return () => {
      scannerRef.current?.stop().catch(() => {});
    };
  }, [navigate]);

  return (
    <div className="aura-backdrop min-h-screen px-4 py-8">
      <div className="max-w-sm mx-auto">
        <button onClick={() => navigate('/dashboard')} className="text-sm text-muted hover:text-rose transition mb-4">
          ← Back
        </button>

        <div className="font-display text-2xl text-ink mb-1">Scan to pay</div>
        <div className="text-sm text-muted mb-6">Point your camera at a PayWave QR code</div>

        <ErrorBanner message={error} />

        <div id={containerId} className="glass-panel rounded-2xl overflow-hidden" />
      </div>
    </div>
  );
}
