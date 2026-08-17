import type { Difficulty } from "../../types";

/** TARGET_SECONDS mirrors internal/scheduler's target durations, for display only. */
export const TARGET_SECONDS: Record<Difficulty, number> = {
  easy: 12 * 60,
  medium: 25 * 60,
  hard: 40 * 60,
};

export const DIFF_LABEL: Record<Difficulty, [label: string, pillClass: string]> = {
  easy: ["Facile", "d-e"],
  medium: ["Moyen", "d-m"],
  hard: ["Difficile", "d-h"],
};

export const FALLBACK_GRADES: Array<[rating: number, label: string, hint: string]> = [
  [1, "Pas résolu", "solution regardée, ou abandon"],
  [2, "Avec effort", "hint utilisé, ou 2× le temps cible"],
  [3, "Résolu", "dans le temps, complexité optimale"],
  [4, "Immédiat", "approche évidente, zéro hésitation"],
];

/** MASTERY_LABEL mirrors internal/scheduler's mastery levels (0..4), for display only. */
export const MASTERY_LABEL = ["Jamais tenté", "À reprendre", "Vu une fois", "Solide", "Maîtrisé"];

export function mmss(totalSeconds: number): string {
  const minutes = Math.floor(totalSeconds / 60)
    .toString()
    .padStart(2, "0");
  const seconds = (totalSeconds % 60).toString().padStart(2, "0");
  return `${minutes}:${seconds}`;
}

export function neetcodeUrl(slug: string): string {
  return `https://neetcode.io/problems/${slug}`;
}

export function leetcodeUrl(slug: string): string {
  return `https://leetcode.com/problems/${slug}/`;
}
