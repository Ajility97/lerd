import { apiJson } from './api';

export type GitDiffSection = 'staged' | 'unstaged' | 'untracked';

export interface GitChangedFile {
  path: string;
  // git's letter: A, M, D, T, U, or ? for untracked.
  status: string;
}

export type GitChanges = Record<GitDiffSection, GitChangedFile[]>;

export interface GitVersions {
  original: string;
  modified: string;
  binary: boolean;
  too_large: boolean;
  unreadable: boolean;
}

export interface GitCommit {
  hash: string;
  subject: string;
  // Unix seconds.
  time: number;
}

// Newest first; more says the server cut a longer list.
export interface GitCommitList {
  commits: GitCommit[];
  more: boolean;
}

// incoming is what a pull brings in, outgoing what a push sends.
export type GitCommitDirection = 'incoming' | 'outgoing';

export const SECTIONS: GitDiffSection[] = ['staged', 'unstaged', 'untracked'];

function query(domain: string, path: string): string {
  return `domain=${encodeURIComponent(domain)}&path=${encodeURIComponent(path)}`;
}

export function loadGitChanges(domain: string, path: string): Promise<GitChanges> {
  return apiJson<GitChanges>(`/api/sites/git-diff?${query(domain, path)}`);
}

export function loadGitVersions(domain: string, path: string, section: GitDiffSection, file: string): Promise<GitVersions> {
  return apiJson<GitVersions>(
    `/api/sites/git-diff?${query(domain, path)}&section=${section}&file=${encodeURIComponent(file)}&content=1`
  );
}

export function loadGitCommits(domain: string, path: string, direction: GitCommitDirection): Promise<GitCommitList> {
  return apiJson<GitCommitList>(`/api/sites/git-diff?${query(domain, path)}&commits=${direction}`);
}

export function loadCommitFiles(domain: string, path: string, hash: string): Promise<GitChangedFile[]> {
  return apiJson<GitChangedFile[]>(`/api/sites/git-diff?${query(domain, path)}&commit=${hash}`);
}

export function loadCommitVersions(domain: string, path: string, hash: string, file: string): Promise<GitVersions> {
  return apiJson<GitVersions>(`/api/sites/git-diff?${query(domain, path)}&commit=${hash}&file=${encodeURIComponent(file)}`);
}

// upto compares HEAD with the commit a pull plan stops at, so the list is
// every file the pull changes, counted the way the plan counts them.
export function loadUptoFiles(domain: string, path: string, hash: string): Promise<GitChangedFile[]> {
  return apiJson<GitChangedFile[]>(`/api/sites/git-diff?${query(domain, path)}&upto=${hash}`);
}

export function loadUptoVersions(domain: string, path: string, hash: string, file: string): Promise<GitVersions> {
  return apiJson<GitVersions>(`/api/sites/git-diff?${query(domain, path)}&upto=${hash}&file=${encodeURIComponent(file)}`);
}

// languageFor picks one of the grammars lerd's Monaco build carries; anything
// else is shown as plain text.
export function languageFor(file: string): string {
  const name = file.split('/').pop() ?? file;
  if (name.endsWith('.php')) return 'php';
  if (name === '.env' || name.startsWith('.env.')) return 'dotenv';
  if (name.endsWith('.ini')) return 'ini';
  if (name.endsWith('.conf')) return 'nginx';
  return 'plaintext';
}
