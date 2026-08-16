import React, { useState, useEffect, useMemo, useCallback } from "react";

/* ---------------------------------------------------------------- data --- */

const E = 1, M = 2, H = 3;

const ROADMAP = [
  ["Arrays & Hashing", [
    ["Contains Duplicate", E], ["Valid Anagram", E], ["Two Sum", E],
    ["Group Anagrams", M], ["Top K Frequent Elements", M],
    ["Encode and Decode Strings", M], ["Product of Array Except Self", M],
    ["Valid Sudoku", M], ["Longest Consecutive Sequence", M],
  ]],
  ["Two Pointers", [
    ["Valid Palindrome", E], ["Two Sum II Input Array Is Sorted", M],
    ["3Sum", M], ["Container With Most Water", M], ["Trapping Rain Water", H],
  ]],
  ["Sliding Window", [
    ["Best Time to Buy and Sell Stock", E],
    ["Longest Substring Without Repeating Characters", M],
    ["Longest Repeating Character Replacement", M],
    ["Permutation in String", M], ["Minimum Window Substring", H],
    ["Sliding Window Maximum", H],
  ]],
  ["Stack", [
    ["Valid Parentheses", E], ["Min Stack", M],
    ["Evaluate Reverse Polish Notation", M], ["Generate Parentheses", M],
    ["Daily Temperatures", M], ["Car Fleet", M],
    ["Largest Rectangle in Histogram", H],
  ]],
  ["Binary Search", [
    ["Binary Search", E], ["Search a 2D Matrix", M], ["Koko Eating Bananas", M],
    ["Find Minimum in Rotated Sorted Array", M],
    ["Search in Rotated Sorted Array", M], ["Time Based Key Value Store", M],
    ["Median of Two Sorted Arrays", H],
  ]],
  ["Linked List", [
    ["Reverse Linked List", E], ["Merge Two Sorted Lists", E],
    ["Reorder List", M], ["Remove Nth Node From End of List", M],
    ["Copy List with Random Pointer", M], ["Add Two Numbers", M],
    ["Linked List Cycle", E], ["Find the Duplicate Number", M],
    ["LRU Cache", M], ["Merge k Sorted Lists", H],
    ["Reverse Nodes in k Group", H],
  ]],
  ["Trees", [
    ["Invert Binary Tree", E], ["Maximum Depth of Binary Tree", E],
    ["Diameter of Binary Tree", E], ["Balanced Binary Tree", E],
    ["Same Tree", E], ["Subtree of Another Tree", E],
    ["Lowest Common Ancestor of a Binary Search Tree", M],
    ["Binary Tree Level Order Traversal", M],
    ["Binary Tree Right Side View", M],
    ["Count Good Nodes in Binary Tree", M],
    ["Validate Binary Search Tree", M],
    ["Kth Smallest Element in a BST", M],
    ["Construct Binary Tree from Preorder and Inorder Traversal", M],
    ["Binary Tree Maximum Path Sum", H],
    ["Serialize and Deserialize Binary Tree", H],
  ]],
  ["Tries", [
    ["Implement Trie Prefix Tree", M],
    ["Design Add and Search Words Data Structure", M],
    ["Word Search II", H],
  ]],
  ["Heap / Priority Queue", [
    ["Kth Largest Element in a Stream", E], ["Last Stone Weight", E],
    ["K Closest Points to Origin", M],
    ["Kth Largest Element in an Array", M], ["Task Scheduler", M],
    ["Design Twitter", M], ["Find Median from Data Stream", H],
  ]],
  ["Backtracking", [
    ["Subsets", M], ["Combination Sum", M], ["Permutations", M],
    ["Subsets II", M], ["Combination Sum II", M], ["Word Search", M],
    ["Palindrome Partitioning", M],
    ["Letter Combinations of a Phone Number", M], ["N Queens", H],
  ]],
  ["Graphs", [
    ["Number of Islands", M], ["Clone Graph", M], ["Max Area of Island", M],
    ["Pacific Atlantic Water Flow", M], ["Surrounded Regions", M],
    ["Rotting Oranges", M], ["Walls and Gates", M], ["Course Schedule", M],
    ["Course Schedule II", M], ["Redundant Connection", M],
    ["Number of Connected Components in an Undirected Graph", M],
    ["Graph Valid Tree", M], ["Word Ladder", H],
  ]],
  ["Advanced Graphs", [
    ["Reconstruct Itinerary", H], ["Min Cost to Connect All Points", M],
    ["Network Delay Time", M], ["Swim in Rising Water", H],
    ["Alien Dictionary", H], ["Cheapest Flights Within K Stops", M],
  ]],
  ["1-D DP", [
    ["Climbing Stairs", E], ["Min Cost Climbing Stairs", E],
    ["House Robber", M], ["House Robber II", M],
    ["Longest Palindromic Substring", M], ["Palindromic Substrings", M],
    ["Decode Ways", M], ["Coin Change", M],
    ["Maximum Product Subarray", M], ["Word Break", M],
    ["Longest Increasing Subsequence", M],
    ["Partition Equal Subset Sum", M],
  ]],
  ["2-D DP", [
    ["Unique Paths", M], ["Longest Common Subsequence", M],
    ["Best Time to Buy and Sell Stock with Cooldown", M],
    ["Coin Change II", M], ["Target Sum", M], ["Interleaving String", M],
    ["Longest Increasing Path in a Matrix", H],
    ["Distinct Subsequences", H], ["Edit Distance", H],
    ["Burst Balloons", H], ["Regular Expression Matching", H],
  ]],
  ["Greedy", [
    ["Maximum Subarray", M], ["Jump Game", M], ["Jump Game II", M],
    ["Gas Station", M], ["Hand of Straights", M],
    ["Merge Triplets to Form Target Triplet", M], ["Partition Labels", M],
    ["Valid Parenthesis String", M],
  ]],
  ["Intervals", [
    ["Insert Interval", M], ["Merge Intervals", M],
    ["Non overlapping Intervals", M], ["Meeting Rooms", E],
    ["Meeting Rooms II", M],
    ["Minimum Interval to Include Each Query", H],
  ]],
  ["Math & Geometry", [
    ["Rotate Image", M], ["Spiral Matrix", M], ["Set Matrix Zeroes", M],
    ["Happy Number", E], ["Plus One", E], ["Pow(x, n)", M],
    ["Multiply Strings", M], ["Detect Squares", M],
  ]],
  ["Bit Manipulation", [
    ["Single Number", E], ["Number of 1 Bits", E], ["Counting Bits", E],
    ["Reverse Bits", E], ["Missing Number", E], ["Sum of Two Integers", M],
    ["Reverse Integer", M],
  ]],
];

const slugify = (t) =>
  t.toLowerCase().replace(/[^a-z0-9 ]/g, "").trim().replace(/\s+/g, "-");

const PROBLEMS = ROADMAP.flatMap(([pattern, list], pi) =>
  list.map(([title, diff], i) => ({
    slug: slugify(title),
    title,
    diff,
    pattern,
    order: pi * 100 + i,
  }))
);

const PATTERNS = ROADMAP.map(([p]) => p);
const BY_PATTERN = Object.fromEntries(
  PATTERNS.map((p) => [p, PROBLEMS.filter((x) => x.pattern === p)])
);

/* ---------------------------------------------------------------- fsrs --- */

const W = [0.4872, 1.4003, 3.7145, 13.8206, 5.1618, 1.2298, 0.8975, 0.031,
  1.6474, 0.1367, 1.0461, 2.1072, 0.0793, 0.3246, 1.587, 0.2272, 2.8755];
const DECAY = -0.5;
const FACTOR = 19 / 81;
const DAY = 86400000;

const clampD = (d) => Math.min(Math.max(d, 1), 10);
const initS = (g) => W[g - 1];
const initD = (g) => clampD(W[4] - (g - 3) * W[5]);

const retrievability = (elapsedDays, S) =>
  Math.pow(1 + (FACTOR * Math.max(elapsedDays, 0)) / S, DECAY);

const intervalFor = (S, req) => (S / FACTOR) * (Math.pow(req, 1 / DECAY) - 1);

function review(card, grade, now, retention) {
  let S, D;
  if (!card) {
    S = initS(grade);
    D = initD(grade);
  } else {
    const elapsed = (now - new Date(card.last).getTime()) / DAY;
    const R = retrievability(elapsed, card.S);
    D = clampD(W[7] * initD(4) + (1 - W[7]) * (card.D - W[6] * (grade - 3)));
    if (grade === 1) {
      const Sf =
        W[11] *
        Math.pow(card.D, -W[12]) *
        (Math.pow(card.S + 1, W[13]) - 1) *
        Math.exp(W[14] * (1 - R));
      S = Math.min(Sf, card.S);
    } else {
      S =
        card.S *
        (1 +
          Math.exp(W[8]) *
            (11 - card.D) *
            Math.pow(card.S, -W[9]) *
            (Math.exp(W[10] * (1 - R)) - 1) *
            (grade === 2 ? W[15] : 1) *
            (grade === 4 ? W[16] : 1));
    }
  }
  S = Math.max(S, 0.1);
  const days = Math.max(1, Math.round(intervalFor(S, retention)));
  return {
    S,
    D,
    last: new Date(now).toISOString(),
    due: new Date(now + days * DAY).toISOString(),
    reps: (card?.reps || 0) + 1,
    lapses: (card?.lapses || 0) + (grade === 1 ? 1 : 0),
    interval: days,
  };
}

const cardR = (card, now) =>
  card ? retrievability((now - new Date(card.last).getTime()) / DAY, card.S) : null;

/* ------------------------------------------------------------- storage --- */

const KEY = "neetrack:v1";
const EMPTY = { log: {}, cards: {}, retention: 0.9, block: 4, links: {}, size: 5, newPct: 40, openMax: 1, name: "" };

async function load() {
  try {
    const r = await window.storage.get(KEY);
    return r ? { ...EMPTY, ...JSON.parse(r.value) } : EMPTY;
  } catch {
    return EMPTY;
  }
}
async function save(state) {
  try {
    await window.storage.set(KEY, JSON.stringify(state));
    return true;
  } catch {
    return false;
  }
}

/* --------------------------------------------------------------- logic --- */

const dayKey = (d) =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;

const HEAT_WEEKS = 26;

function activity(log) {
  const counts = {};
  Object.values(log).forEach((arr) =>
    arr.forEach((a) => {
      const k = dayKey(new Date(a.t));
      counts[k] = (counts[k] || 0) + 1;
    })
  );
  return counts;
}

function streakOf(counts) {
  const d = new Date();
  d.setHours(0, 0, 0, 0);
  if (!counts[dayKey(d)]) d.setDate(d.getDate() - 1);
  let s = 0;
  while (counts[dayKey(d)]) {
    s += 1;
    d.setDate(d.getDate() - 1);
  }
  return s;
}

function heatWeeks(counts) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const end = new Date(today);
  end.setDate(end.getDate() + (6 - ((today.getDay() + 6) % 7)));
  const cursor = new Date(end);
  cursor.setDate(cursor.getDate() - (HEAT_WEEKS * 7 - 1));
  const weeks = [];
  for (let w = 0; w < HEAT_WEEKS; w += 1) {
    const days = [];
    for (let d = 0; d < 7; d += 1) {
      const date = new Date(cursor);
      const k = dayKey(date);
      days.push({
        key: k,
        date,
        future: date > today,
        n: counts[k] || 0,
      });
      cursor.setDate(cursor.getDate() + 1);
    }
    weeks.push(days);
  }
  return weeks;
}

const heatLevel = (n) => (n === 0 ? 0 : n === 1 ? 1 : n <= 2 ? 2 : n <= 4 ? 3 : 4);

const DIFF = { 1: ["Facile", "d-e"], 2: ["Moyen", "d-m"], 3: ["Difficile", "d-h"] };
const TARGET = { 1: 12 * 60, 2: 25 * 60, 3: 40 * 60 };
const mmss = (s) =>
  `${Math.floor(s / 60).toString().padStart(2, "0")}:${(s % 60).toString().padStart(2, "0")}`;
const GRADES = [
  [1, "Pas résolu", "solution regardée, ou abandon"],
  [2, "Avec effort", "hint utilisé, ou 2× le temps cible"],
  [3, "Résolu", "dans le temps, complexité optimale"],
  [4, "Immédiat", "approche évidente, zéro hésitation"],
];

function mastery(attempts) {
  if (!attempts || !attempts.length) return 0;
  const last = attempts[attempts.length - 1].g;
  const solves = attempts.filter((a) => a.g >= 3).length;
  if (last === 1) return 1;
  if (solves >= 3 && last === 4) return 4;
  if (solves >= 2) return 3;
  return 2;
}
const MASTERY_LABEL = ["Jamais tenté", "À reprendre", "Vu une fois", "Solide", "Maîtrisé"];

function composeSession(state, n, now, ratio) {
  const picked = [];
  const used = new Set();

  const due = PATTERNS.map((p) => ({ p, card: state.cards[p] }))
    .filter((x) => x.card && new Date(x.card.due).getTime() <= now)
    .map((x) => ({ ...x, r: cardR(x.card, now) }))
    .sort((a, b) => a.r - b.r);

  const reviewQuota = due.length ? Math.round(n * (1 - ratio)) : 0;

  const pickFrom = (pattern) => {
    const pool = BY_PATTERN[pattern].filter((x) => !used.has(x.slug));
    if (!pool.length) return null;
    const fresh = pool.filter((x) => !(state.log[x.slug] || []).length);
    if (fresh.length) return fresh[0];
    return pool.sort((a, b) => {
      const la = state.log[a.slug], lb = state.log[b.slug];
      return la[la.length - 1].t - lb[lb.length - 1].t;
    })[0];
  };

  for (const d of due) {
    if (picked.length >= reviewQuota) break;
    const prob = pickFrom(d.p);
    if (prob) {
      used.add(prob.slug);
      picked.push({ ...prob, why: "review", r: d.r });
    }
  }

  const perPattern = {};
  let opened = 0;
  const block = state.block || 4;
  const fresh = PROBLEMS.filter((x) => !used.has(x.slug) && !(state.log[x.slug] || []).length)
    .sort((a, b) => a.order - b.order);
  for (const prob of fresh) {
    if (picked.length >= n) break;
    const known = !!state.cards[prob.pattern];
    const taken = perPattern[prob.pattern] || 0;
    if (taken >= (known ? 1 : block)) continue;
    if (!known && taken === 0) {
      if (opened >= (state.openMax || 1)) continue;
      opened += 1;
    }
    perPattern[prob.pattern] = taken + 1;
    used.add(prob.slug);
    picked.push({ ...prob, why: "new" });
  }

  for (const d of due) {
    if (picked.length >= n) break;
    const prob = pickFrom(d.p);
    if (prob) {
      used.add(prob.slug);
      picked.push({ ...prob, why: "review", r: d.r });
    }
  }

  return picked.sort((a, b) => (a.why === b.why ? 0 : a.why === "review" ? -1 : 1));
}

/* ------------------------------------------------------------------ ui --- */

const CSS = `
.nt { --ink:#141C26; --ink-2:#5A6875; --ink-3:#8C97A2; --paper:#EDF0F3;
  --card:#FFFFFF; --rule:#D8DEE4; --sig:#C11F44; --ok:#0E7C6B;
  --new:#3355C9;
  background:var(--paper); color:var(--ink); min-height:100vh;
  font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;
  font-size:15px; line-height:1.5; }
.nt *,.nt *::before,.nt *::after { box-sizing:border-box; }
.nt .mono { font-family:ui-monospace,SFMono-Regular,Menlo,"Cascadia Mono",monospace; }
.nt .wrap { max-width:940px; margin:0 auto; padding:28px 20px 80px; }
.nt .eyebrow { font-size:11px; letter-spacing:.16em; text-transform:uppercase;
  color:var(--ink-3); font-family:ui-monospace,SFMono-Regular,Menlo,monospace; }
.nt h1 { font-size:26px; font-weight:600; letter-spacing:-.02em; margin:6px 0 0; }
.nt .stats { display:flex; gap:28px; flex-wrap:wrap; margin-top:18px;
  padding-bottom:22px; border-bottom:1px solid var(--rule); }
.nt .stat b { display:block; font-size:30px; font-weight:600; letter-spacing:-.03em;
  font-family:ui-monospace,SFMono-Regular,Menlo,monospace; line-height:1.1; }
.nt .stat span { font-size:11px; letter-spacing:.1em; text-transform:uppercase;
  color:var(--ink-3); font-family:ui-monospace,SFMono-Regular,Menlo,monospace; }
.nt .top { display:flex; align-items:flex-start; justify-content:space-between; gap:16px; }
.nt .who { background:none; border:1px solid var(--rule); border-radius:999px;
  padding:6px 14px; font:inherit; font-size:13px; color:var(--ink-2); cursor:pointer;
  white-space:nowrap; flex-shrink:0; }
.nt .who:hover { border-color:var(--ink); color:var(--ink); }
.nt .who b { font-weight:500; color:var(--ink); }
.nt .setup { display:flex; align-items:center; gap:12px; flex-wrap:wrap; }
.nt .setup label { font-size:13px; color:var(--ink-2); min-width:120px; }
.nt .tabs { display:flex; gap:2px; margin:22px 0 20px; }
.nt .heat { padding:18px 0 4px; border-bottom:1px solid var(--rule); }
.nt .heat-scroll { overflow-x:auto; padding-bottom:6px; }
.nt .heat-inner { display:inline-block; min-width:100%; }
.nt .heat-months { display:flex; gap:3px; margin-bottom:5px; padding-left:22px; }
.nt .heat-months span { width:11px; flex:0 0 11px; font-size:10px; color:var(--ink-3);
  font-family:ui-monospace,monospace; white-space:nowrap; overflow:visible; }
.nt .heat-body { display:flex; gap:3px; }
.nt .heat-days { display:flex; flex-direction:column; gap:3px; width:19px; flex:0 0 19px; }
.nt .heat-days span { height:11px; font-size:9px; line-height:11px; color:var(--ink-3);
  font-family:ui-monospace,monospace; }
.nt .heat-col { display:flex; flex-direction:column; gap:3px; }
.nt .heat-col i { width:11px; height:11px; border-radius:2px; display:block;
  background:#E4E9ED; }
.nt .heat-col i.l1 { background:#C7E4DD; }
.nt .heat-col i.l2 { background:#8ACCBC; }
.nt .heat-col i.l3 { background:#3FA391; }
.nt .heat-col i.l4 { background:#0E7C6B; }
.nt .heat-col i.void { background:none; }
.nt .heat-foot { display:flex; align-items:center; gap:6px; margin-top:8px;
  font-size:11px; color:var(--ink-3); font-family:ui-monospace,monospace; }
.nt .heat-foot i { width:11px; height:11px; border-radius:2px; display:block; background:#E4E9ED; }
.nt .tab { border:none; background:none; padding:7px 14px; cursor:pointer;
  font-size:13px; color:var(--ink-2); border-radius:999px; font-weight:500; }
.nt .tab[aria-selected="true"] { background:var(--ink); color:var(--paper); }
.nt .tab:focus-visible,.nt button:focus-visible,.nt a:focus-visible,.nt input:focus-visible
  { outline:2px solid var(--new); outline-offset:2px; }
.nt .row { display:flex; align-items:center; gap:14px; padding:11px 14px;
  background:var(--card); border:1px solid var(--rule); border-radius:8px;
  margin-bottom:6px; cursor:pointer; text-align:left; width:100%; font:inherit;
  color:inherit; }
.nt .row:hover { border-color:var(--ink-3); }
.nt .pname { flex:1; font-weight:500; font-size:14px; }
.nt .bar { width:170px; height:7px; background:#E4E9ED; border-radius:4px;
  overflow:hidden; flex-shrink:0; }
.nt .bar i { display:block; height:100%; border-radius:4px; }
.nt .bar.empty { background:none; border:1px dashed var(--rule); height:7px; }
.nt .pct { width:42px; text-align:right; font-size:12px; color:var(--ink-2); }
.nt .count { width:46px; text-align:right; font-size:12px; color:var(--ink-3); }
.nt .kids { padding:2px 0 12px 14px; }
.nt .kid { display:flex; align-items:center; gap:12px; padding:7px 12px;
  border-left:2px solid var(--rule); font-size:13px; }
.nt .kid a { color:inherit; text-decoration:none; flex:1; }
.nt .kid a:hover { text-decoration:underline; }
.nt .dots { display:flex; gap:3px; }
.nt .dots i { width:7px; height:7px; border-radius:50%; background:#DDE3E8; }
.nt .dots i.on { background:var(--ok); }
.nt .dots i.warn { background:var(--sig); }
.nt .pill { font-size:10px; letter-spacing:.08em; text-transform:uppercase;
  padding:2px 7px; border-radius:4px; font-family:ui-monospace,SFMono-Regular,monospace; }
.nt .d-e { background:#E2F0EC; color:#0E7C6B; }
.nt .d-m { background:#FBEEDB; color:#8A5A05; }
.nt .d-h { background:#FAE3E8; color:#A81839; }
.nt .panel { background:var(--card); border:1px solid var(--rule);
  border-radius:10px; padding:20px; }
.nt .btn { background:var(--ink); color:var(--paper); border:none; border-radius:7px;
  padding:10px 18px; font-size:14px; font-weight:500; cursor:pointer; font-family:inherit; }
.nt .btn:hover { background:#2A3542; }
.nt .btn.ghost { background:none; color:var(--ink-2); border:1px solid var(--rule); }
.nt .btn.ghost:hover { background:#F5F7F9; }
.nt .num { width:64px; padding:9px 10px; border:1px solid var(--rule);
  border-radius:7px; font:inherit; font-family:ui-monospace,monospace; }
.nt .task { background:var(--card); border:1px solid var(--rule);
  border-radius:10px; padding:16px 18px; margin-bottom:10px; }
.nt .task.done { opacity:.45; }
.nt .tag { font-size:10px; letter-spacing:.1em; text-transform:uppercase;
  font-family:ui-monospace,monospace; }
.nt .tag.review { color:var(--sig); }
.nt .tag.new { color:var(--new); }
.nt .grades { display:flex; gap:6px; flex-wrap:wrap; margin-top:12px; }
.nt .grade { flex:1; min-width:112px; border:1px solid var(--rule); background:none;
  border-radius:7px; padding:8px 10px; cursor:pointer; text-align:left; font:inherit; }
.nt .grade:hover { border-color:var(--ink); background:#F5F7F9; }
.nt .grade.hint { border-color:var(--sig); }
.nt .grade.hint b::after { content:" · suggéré"; color:var(--sig); font-weight:400; font-size:11px; }
.nt .clock { display:flex; align-items:center; gap:10px; margin-top:10px;
  padding-top:10px; border-top:1px solid var(--rule); flex-wrap:wrap; }
.nt .clock b { font-family:ui-monospace,SFMono-Regular,Menlo,monospace;
  font-size:22px; font-weight:500; font-variant-numeric:tabular-nums;
  letter-spacing:-.02em; }
.nt .clock b.over { color:var(--sig); }
.nt .mini { border:1px solid var(--rule); background:none; border-radius:6px;
  padding:4px 10px; font-size:12px; cursor:pointer; font:inherit; font-size:12px;
  color:var(--ink-2); }
.nt .mini:hover { border-color:var(--ink); color:var(--ink); }
.nt .lc { font-size:11px; color:var(--ink-3); text-decoration:none;
  font-family:ui-monospace,monospace; }
.nt .lc:hover { color:var(--ink); text-decoration:underline; }
.nt .grade b { display:block; font-size:13px; font-weight:500; }
.nt .grade span { font-size:11px; color:var(--ink-3); line-height:1.35; display:block; }
.nt .muted { color:var(--ink-2); font-size:13px; }
.nt .empty-state { text-align:center; padding:44px 20px; color:var(--ink-2); }
@media (max-width:640px) {
  .nt .bar { width:80px; } .nt .stats { gap:18px; }
  .nt .grade { min-width:calc(50% - 3px); }
}
`;

export default function NeetTrack() {
  const [state, setState] = useState(null);
  const [tab, setTab] = useState("patterns");
  const [open, setOpen] = useState(null);
  const [size, setSize] = useState(5);
  const [session, setSession] = useState(null);
  const [now, setNow] = useState(Date.now());
  const [timers, setTimers] = useState({});
  const [, setTick] = useState(0);

  useEffect(() => { load().then((s) => { setState(s); setSize(s.size || 5); }); }, []);

  useEffect(() => {
    if (!Object.values(timers).some((t) => t.start)) return;
    const id = setInterval(() => setTick((x) => x + 1), 1000);
    return () => clearInterval(id);
  }, [timers]);

  const commit = useCallback((next) => {
    setState(next);
    save(next);
  }, []);

  const solved = useMemo(
    () => (state ? PROBLEMS.filter((p) => mastery(state.log[p.slug]) >= 2).length : 0),
    [state]
  );

  const counts = useMemo(() => (state ? activity(state.log) : {}), [state]);
  const weeks = useMemo(() => heatWeeks(counts), [counts]);
  const streak = useMemo(() => streakOf(counts), [counts]);

  const patternRows = useMemo(() => {
    if (!state) return [];
    return PATTERNS.map((p) => {
      const card = state.cards[p];
      const r = cardR(card, now);
      const done = BY_PATTERN[p].filter((x) => mastery(state.log[x.slug]) >= 2).length;
      const isDue = card && new Date(card.due).getTime() <= now;
      return { p, card, r, done, total: BY_PATTERN[p].length, isDue };
    });
  }, [state, now]);

  const dueCount = patternRows.filter((x) => x.isDue).length;

  const preview = useMemo(() => {
    if (!state || session) return null;
    const n = Math.max(1, Math.min(20, size));
    const s = composeSession(state, n, now, (state.newPct ?? 40) / 100);
    return {
      review: s.filter((x) => x.why === "review").length,
      fresh: s.filter((x) => x.why === "new").length,
      total: s.length,
      asked: n,
    };
  }, [state, size, session, now]);

  if (!state) {
    return (
      <div className="nt">
        <style>{CSS}</style>
        <div className="wrap"><p className="muted">Chargement…</p></div>
      </div>
    );
  }

  const start = () => {
    const t = Date.now();
    const n = Math.max(1, Math.min(20, size));
    setNow(t);
    setTimers({});
    commit({ ...state, size: n });
    setSession(composeSession(state, n, t, (state.newPct ?? 40) / 100).map((x) => ({ ...x, done: null })));
    setTab("session");
  };

  const secondsOf = (slug) => {
    const t = timers[slug];
    if (!t) return 0;
    return t.acc + (t.start ? Math.floor((Date.now() - t.start) / 1000) : 0);
  };

  const toggleTimer = (slug) =>
    setTimers((prev) => {
      const t = prev[slug] || { acc: 0, start: null };
      return {
        ...prev,
        [slug]: t.start
          ? { acc: t.acc + Math.floor((Date.now() - t.start) / 1000), start: null }
          : { acc: t.acc, start: Date.now() },
      };
    });

  const ncUrl = (slug) => `https://neetcode.io/problems/${state.links[slug] || slug}`;

  const fixLink = (slug) => {
    const v = prompt(
      "Slug ou URL NeetCode pour ce problème",
      state.links[slug] || slug
    );
    if (!v) return;
    const clean = v.trim().replace(/^.*\/problems\//, "").replace(/\/+$/, "");
    if (clean) commit({ ...state, links: { ...state.links, [slug]: clean } });
  };

  const grade = (idx, g) => {
    const item = session[idx];
    const t = Date.now();
    const secs = secondsOf(item.slug);
    setTimers((prev) => ({ ...prev, [item.slug]: { acc: secs, start: null } }));
    const entry = secs ? { t, g, s: secs } : { t, g };
    const next = {
      ...state,
      log: { ...state.log, [item.slug]: [...(state.log[item.slug] || []), entry] },
      cards: { ...state.cards, [item.pattern]: review(state.cards[item.pattern], g, t, state.retention) },
    };
    commit(next);
    setSession(session.map((s, i) => (i === idx ? { ...s, done: g, secs } : s)));
  };

  const remaining = session ? session.filter((s) => s.done === null).length : 0;

  return (
    <div className="nt">
      <style>{CSS}</style>
      <div className="wrap">
        <div className="top">
          <div>
            <div className="eyebrow">Neetcode 150 · rétention espacée</div>
            <h1>{state.name ? `Le suivi de ${state.name}` : "Suivi et sessions"}</h1>
          </div>
          <button className="who" onClick={() => {
            const v = prompt("Ton pseudo", state.name || "");
            if (v !== null) commit({ ...state, name: v.trim().slice(0, 24) });
          }}>
            {state.name ? <b>{state.name}</b> : "Ajouter un pseudo"}
          </button>
        </div>

        <div className="stats">
          <div className="stat">
            <b>{solved}<span style={{ fontSize: 16, color: "var(--ink-3)" }}>/150</span></b>
            <span>Problèmes vus</span>
          </div>
          <div className="stat">
            <b>{patternRows.filter((x) => x.card).length}<span style={{ fontSize: 16, color: "var(--ink-3)" }}>/{PATTERNS.length}</span></b>
            <span>Patterns amorcés</span>
          </div>
          <div className="stat">
            <b style={{ color: dueCount ? "var(--sig)" : undefined }}>{dueCount}</b>
            <span>Patterns à revoir</span>
          </div>
          <div className="stat">
            <b>{streak}<span style={{ fontSize: 16, color: "var(--ink-3)" }}> j</span></b>
            <span>Série en cours</span>
          </div>
        </div>

        <div className="heat">
          <div className="heat-scroll">
            <div className="heat-inner">
              <div className="heat-months">
                {weeks.map((w, i) => {
                  const m = w[0].date.getMonth();
                  const prev = i ? weeks[i - 1][0].date.getMonth() : -1;
                  return (
                    <span key={i}>
                      {m !== prev && i < weeks.length - 1
                        ? w[0].date.toLocaleDateString("fr-FR", { month: "short" }).replace(".", "")
                        : ""}
                    </span>
                  );
                })}
              </div>
              <div className="heat-body">
                <div className="heat-days">
                  <span>lun</span><span /><span>mer</span><span /><span>ven</span><span /><span />
                </div>
                {weeks.map((w, i) => (
                  <div className="heat-col" key={i}>
                    {w.map((d) => (
                      <i
                        key={d.key}
                        className={d.future ? "void" : `l${heatLevel(d.n)}`}
                        title={
                          d.future
                            ? ""
                            : `${d.n} exercice${d.n > 1 ? "s" : ""} · ${d.date.toLocaleDateString("fr-FR", { day: "numeric", month: "long" })}`
                        }
                      />
                    ))}
                  </div>
                ))}
              </div>
            </div>
          </div>
          <div className="heat-foot">
            <span>Moins</span>
            <i /><i style={{ background: "#C7E4DD" }} /><i style={{ background: "#8ACCBC" }} />
            <i style={{ background: "#3FA391" }} /><i style={{ background: "#0E7C6B" }} />
            <span>Plus</span>
            <span style={{ marginLeft: "auto" }}>{HEAT_WEEKS} dernières semaines</span>
          </div>
        </div>

        <div className="tabs" role="tablist">
          <button className="tab" role="tab" aria-selected={tab === "patterns"} onClick={() => setTab("patterns")}>Patterns</button>
          <button className="tab" role="tab" aria-selected={tab === "session"} onClick={() => setTab("session")}>Session</button>
          <button className="tab" role="tab" aria-selected={tab === "settings"} onClick={() => setTab("settings")}>Réglages</button>
        </div>

        {tab === "patterns" && (
          <div>
            {patternRows.map(({ p, card, r, done, total, isDue }) => (
              <div key={p}>
                <button className="row" onClick={() => setOpen(open === p ? null : p)} aria-expanded={open === p}>
                  <span className="pname">{p}</span>
                  {card ? (
                    <span className="bar">
                      <i style={{ width: `${Math.round(r * 100)}%`, background: isDue ? "var(--sig)" : "var(--ok)" }} />
                    </span>
                  ) : (
                    <span className="bar empty" />
                  )}
                  <span className="pct mono">{card ? `${Math.round(r * 100)}%` : "—"}</span>
                  <span className="count mono">{done}/{total}</span>
                </button>
                {open === p && (
                  <div className="kids">
                    {BY_PATTERN[p].map((prob) => {
                      const m = mastery(state.log[prob.slug]);
                      return (
                        <div className="kid" key={prob.slug}>
                          <span className="dots" title={MASTERY_LABEL[m]}>
                            {[1, 2, 3, 4].map((i) => (
                              <i key={i} className={m === 1 && i === 1 ? "warn" : i <= m ? "on" : ""} />
                            ))}
                          </span>
                          <a href={ncUrl(prob.slug)} target="_blank" rel="noreferrer">{prob.title}</a>
                          <button className="lc" onClick={() => fixLink(prob.slug)} title="Corriger le lien NeetCode">lien</button>
                          <a className="lc" href={`https://leetcode.com/problems/${prob.slug}/`} target="_blank" rel="noreferrer">LC</a>
                          <span className={`pill ${DIFF[prob.diff][1]}`}>{DIFF[prob.diff][0]}</span>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}

        {tab === "session" && (
          <div>
            {!session && (
              <div className="panel">
                <div className="setup">
                  <label htmlFor="n">Exercices</label>
                  <input id="n" className="num mono" type="number" min="1" max="20" value={size}
                    onChange={(e) => setSize(Number(e.target.value))} />
                </div>
                <div className="setup" style={{ marginTop: 16 }}>
                  <label htmlFor="mix">Part de nouveaux</label>
                  <input id="mix" type="range" min="0" max="100" step="10"
                    value={state.newPct ?? 40}
                    onChange={(e) => commit({ ...state, newPct: Number(e.target.value) })}
                    style={{ flex: 1, minWidth: 140 }}
                    disabled={!dueCount} />
                  <span className="mono" style={{ width: 44, textAlign: "right" }}>{state.newPct ?? 40}%</span>
                </div>
                <p className="muted" style={{ margin: "14px 0 0" }}>
                  {!dueCount
                    ? "Rien à revoir aujourd'hui — la session sera entièrement composée de nouveaux problèmes."
                    : preview && (
                      <>
                        <span className="mono" style={{ color: "var(--sig)" }}>{preview.review} révision{preview.review > 1 ? "s" : ""}</span>
                        {" · "}
                        <span className="mono" style={{ color: "var(--new)" }}>{preview.fresh} nouveau{preview.fresh > 1 ? "x" : ""}</span>
                        {preview.total < preview.asked && ` — seulement ${preview.total} disponibles`}
                      </>
                    )}
                </p>
                <button className="btn" style={{ marginTop: 16 }} onClick={start}>Composer la session</button>
              </div>
            )}

            {session && (
              <div>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14 }}>
                  <span className="muted mono">{session.length - remaining}/{session.length} noté{session.length - remaining > 1 ? "s" : ""}</span>
                  <button className="btn ghost" onClick={() => setSession(null)}>
                    {remaining ? "Abandonner" : "Terminer"}
                  </button>
                </div>
                {session.map((item, i) => (
                  <div className={`task${item.done ? " done" : ""}`} key={item.slug}>
                    <div style={{ display: "flex", alignItems: "baseline", gap: 10, flexWrap: "wrap" }}>
                      <span className={`tag ${item.why}`}>
                        {item.why === "review" ? `Révision · ${Math.round(item.r * 100)}%` : "Nouveau"}
                      </span>
                      <span className="muted" style={{ fontSize: 12 }}>{item.pattern}</span>
                      <span className={`pill ${DIFF[item.diff][1]}`}>{DIFF[item.diff][0]}</span>
                    </div>
                    <div style={{ display: "flex", alignItems: "baseline", gap: 10, marginTop: 6, flexWrap: "wrap" }}>
                      <a href={ncUrl(item.slug)} target="_blank" rel="noreferrer"
                        style={{ fontSize: 16, fontWeight: 500, color: "inherit" }}>
                        {item.title} ↗
                      </a>
                      <button className="lc" onClick={() => fixLink(item.slug)}>lien</button>
                      <a className="lc" href={`https://leetcode.com/problems/${item.slug}/`} target="_blank" rel="noreferrer">LC</a>
                    </div>
                    {item.done ? (
                      <p className="muted" style={{ margin: "10px 0 0" }}>
                        {item.secs ? `${mmss(item.secs)} · ` : ""}
                        noté « {GRADES.find((g) => g[0] === item.done)[1]} » · {item.pattern} revient dans {state.cards[item.pattern].interval} j
                      </p>
                    ) : (
                      <>
                        <div className="clock">
                          <b className={secondsOf(item.slug) > TARGET[item.diff] ? "over" : ""}>
                            {mmss(secondsOf(item.slug))}
                          </b>
                          <button className="mini" onClick={() => toggleTimer(item.slug)}>
                            {timers[item.slug]?.start ? "Pause" : secondsOf(item.slug) ? "Reprendre" : "Démarrer"}
                          </button>
                          <span className="muted" style={{ fontSize: 12 }}>cible {mmss(TARGET[item.diff])}</span>
                        </div>
                        <div className="grades">
                          {GRADES.map(([g, label, hint]) => {
                            const s = secondsOf(item.slug);
                            const suggest = s > 2 * TARGET[item.diff] ? 2 : s > TARGET[item.diff] ? 2 : s ? 3 : 0;
                            return (
                              <button className={`grade${suggest === g ? " hint" : ""}`} key={g} onClick={() => grade(i, g)}>
                                <b>{label}</b><span>{hint}</span>
                              </button>
                            );
                          })}
                        </div>
                      </>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {tab === "settings" && (
          <div className="panel">
            <label htmlFor="ret" style={{ fontWeight: 500 }}>Rétention cible</label>
            <p className="muted" style={{ marginTop: 4 }}>
              La probabilité de réussite visée au moment où un pattern revient. Plus haut = intervalles courts et charge lourde.
              Monte vers 0.95 à l'approche des entretiens.
            </p>
            <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
              <input id="ret" type="range" min="80" max="97" value={Math.round(state.retention * 100)}
                onChange={(e) => commit({ ...state, retention: Number(e.target.value) / 100 })}
                style={{ flex: 1 }} />
              <span className="mono" style={{ width: 46, textAlign: "right" }}>{state.retention.toFixed(2)}</span>
            </div>
            <hr style={{ border: "none", borderTop: "1px solid var(--rule)", margin: "22px 0" }} />
            <label htmlFor="blk" style={{ fontWeight: 500 }}>Taille du bloc d'apprentissage</label>
            <p className="muted" style={{ marginTop: 4 }}>
              Quand un pattern est découvert, il est servi en bloc : plusieurs problèmes d'affilée pour laisser
              le schéma se former. Une session n'ouvre jamais plus d'un pattern neuf. Les patterns déjà connus
              sont ensuite entrelacés, un problème à la fois.
            </p>
            <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
              <input id="blk" type="range" min="1" max="6" value={state.block || 4}
                onChange={(e) => commit({ ...state, block: Number(e.target.value) })}
                style={{ flex: 1 }} />
              <span className="mono" style={{ width: 46, textAlign: "right" }}>{state.block || 4}</span>
            </div>
            <hr style={{ border: "none", borderTop: "1px solid var(--rule)", margin: "22px 0" }} />
            <button className="btn ghost"
              onClick={() => { if (confirm("Effacer tout l'historique et les cartes ?")) commit(EMPTY); }}>
              Effacer les données
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
