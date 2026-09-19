import { getInitials, getAvatarColor } from '../utils/avatar';

export function Avatar({ name, size = 40 }) {
  return (
    <div
      className="rounded-full flex items-center justify-center text-white font-display font-medium shrink-0"
      style={{
        width: size,
        height: size,
        fontSize: size * 0.4,
        backgroundColor: getAvatarColor(name),
      }}
    >
      {getInitials(name)}
    </div>
  );
}
