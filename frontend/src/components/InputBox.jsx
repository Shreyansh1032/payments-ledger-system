export function InputBox({ label, placeholder, onChange, type = 'text', value }) {
  return (
    <div>
      <div className="text-sm font-medium text-left py-2 text-ink/80">
        {label}
      </div>
      <input
        type={type}
        placeholder={placeholder}
        value={value}
        className="w-full px-3 py-2 border border-ink/10 bg-white/70 rounded-xl text-sm placeholder:text-muted focus:outline-none focus:ring-2 focus:ring-rose/30 focus:border-rose/40 transition"
        onChange={onChange}
      />
    </div>
  );
}
