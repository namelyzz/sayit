const COMMUNITY_PALETTES = [
  { bg: "rgba(156, 171, 169, 0.14)", text: "#42504e", border: "rgba(156, 171, 169, 0.32)", dot: "#9caba9" },
  { bg: "rgba(156, 171, 169, 0.18)", text: "#3c4a48", border: "rgba(156, 171, 169, 0.36)", dot: "#8f9f9c" },
  { bg: "rgba(156, 171, 169, 0.22)", text: "#354340", border: "rgba(156, 171, 169, 0.40)", dot: "#82918f" },
];

function hashString(value: string) {
  let hash = 0;
  for (let index = 0; index < value.length; index += 1) {
    hash = (hash << 5) - hash + value.charCodeAt(index);
    hash |= 0;
  }
  return Math.abs(hash);
}

export function getCommunityPalette(seed?: string) {
  if (!seed) return COMMUNITY_PALETTES[0];
  return COMMUNITY_PALETTES[hashString(seed) % COMMUNITY_PALETTES.length];
}
