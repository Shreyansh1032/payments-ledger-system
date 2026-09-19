import { Link } from 'react-router-dom';

export function BottomWarning({ label, buttonText, to }) {
  return (
    <div className="py-2 text-sm flex justify-center gap-1 text-muted font-sans">
      <div>{label}</div>
      <Link className="text-rose font-medium hover:underline" to={to}>
        {buttonText}
      </Link>
    </div>
  );
}
