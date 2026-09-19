const PALETTE = ['#B3245C', '#D6567F', '#7A1F44', '#C9628B', '#E8A0BB'];

export function getInitials(name) {
  if (!name) return '?';
  return name.trim().charAt(0).toUpperCase();
}

export function getAvatarColor(name) {
  if (!name) return PALETTE[0];
  const sum = name.split('').reduce((acc, c) => acc + c.charCodeAt(0), 0);
  return PALETTE[sum % PALETTE.length];
}
