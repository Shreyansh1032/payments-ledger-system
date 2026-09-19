export function Button({ label, onClick, variant = 'primary', type = 'button', disabled = false }) {
  const base = 'w-full font-sans font-semibold rounded-xl text-sm px-5 py-2.5 transition-colors disabled:opacity-50 disabled:cursor-not-allowed';
  const variants = {
    primary: 'text-white bg-rose hover:bg-rose-dark',
    ghost: 'text-rose bg-transparent border border-rose/30 hover:bg-blush',
  };

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className={`${base} ${variants[variant]}`}
    >
      {label}
    </button>
  );
}
