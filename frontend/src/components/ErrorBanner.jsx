export function ErrorBanner({ message }) {
  if (!message) return null;
  return (
    <div className="bg-rose/10 border border-rose/30 text-rose-dark text-sm rounded-xl px-4 py-2.5 mb-4">
      {message}
    </div>
  );
}
