/**
 * data.ts — v1.0 数据加载层
 *
 * Astro SSG 构建时直接读取 data/ 目录的 JSON 文件，
 * 替代 v0.x 的 HTTP API 调用。所有函数仅在 build-time (Node) 执行。
 */
import fs from 'node:fs';
import path from 'node:path';

// 数据目录：默认 ../data（相对于 web/），可通过 DATA_DIR 环境变量覆盖
const DATA_DIR = path.resolve(process.env.DATA_DIR || path.join(process.cwd(), '..', 'data'));

/* ================================================================
 * Types — 严格对齐 Go datastore/models.go
 * ================================================================ */

export interface Trending {
    daily_stars?: number;
    weekly_stars?: number;
    rank_daily?: number;
    last_seen_trending?: string;
}

export interface Feature {
    name: string;
    desc: string;
}

export interface ComparisonEntry {
    project: string;
    diff: string;
}

export interface Analysis {
    status: string;       // draft | published | rejected
    model: string;
    summary: string;
    positioning?: string;
    features?: Feature[];
    advantages?: string;
    tech_stack?: string;
    use_cases?: string;
    comparison?: ComparisonEntry[];
    ecosystem?: string;
    generated_at: string;
    reviewed_at?: string;
    token_usage?: number;
}

export interface CategoryMatch {
    slug: string;
    confidence: number;
}

export interface Project {
    id: string;           // owner__repo
    full_name: string;    // owner/repo
    description?: string;
    language?: string;
    license?: string;
    topics?: string[];
    homepage?: string;
    stars: number;
    forks: number;
    open_issues: number;
    watchers: number;
    is_archived: boolean;
    pushed_at?: string;
    created_at_gh?: string;
    score: number;
    rank?: number;
    category?: string;    // primary category slug
    trending?: Trending;
    analysis?: Analysis;
    categories?: CategoryMatch[];
    first_seen_at: string;
    updated_at: string;
}

export interface Snapshot {
    project_id: string;
    date: string;
    stars: number;
    forks: number;
    open_issues: number;
    watchers?: number;
    score?: number;
    rank?: number;
    daily_stars?: number;
}

export interface RankingItem {
    rank: number;
    project_id: string;
    full_name: string;
    summary?: string;
    language?: string;
    category?: string;
    stars: number;
    daily_stars?: number;
    weekly_stars?: number;
    score: number;
    rank_change?: number;  // positive=up, negative=down, undefined=new
}

export interface Ranking {
    date: string;
    total: number;
    items: RankingItem[];
}

export interface Post {
    slug: string;
    title: string;
    content: string;       // Markdown
    post_type: string;     // weekly | monthly | spotlight
    cover_image_url?: string;
    published_at?: string;
    created_at: string;
    updated_at?: string;
}

export interface CategoryKeywords {
    topics: string[];
    description: string[];
}

export interface Category {
    slug: string;
    name: string;
    description: string;
    sort_order: number;
    keywords: CategoryKeywords;
    project_ids: string[];
}

/* ================================================================
 * Helpers
 * ================================================================ */

function safeReadJSON<T>(filePath: string): T | null {
    try {
        return JSON.parse(fs.readFileSync(filePath, 'utf-8'));
    } catch {
        return null;
    }
}

function safeReadDir(dirPath: string): string[] {
    try {
        return fs.readdirSync(dirPath);
    } catch {
        return [];
    }
}

/* ================================================================
 * Data Access — 构建时执行
 * ================================================================ */

/** 获取最新排行榜（最近日期）。 */
export function getLatestRanking(): Ranking | null {
    const dir = path.join(DATA_DIR, 'rankings');
    const files = safeReadDir(dir).filter(f => f.endsWith('.json')).sort().reverse();
    if (files.length === 0) return null;
    return safeReadJSON<Ranking>(path.join(dir, files[0]));
}

/** 获取指定 ID 的项目（owner__repo）。 */
export function getProject(id: string): Project | null {
    return safeReadJSON<Project>(path.join(DATA_DIR, 'projects', `${id}.json`));
}

/** 获取所有项目。 */
export function getAllProjects(): Project[] {
    const dir = path.join(DATA_DIR, 'projects');
    return safeReadDir(dir)
        .filter(f => f.endsWith('.json'))
        .map(f => safeReadJSON<Project>(path.join(dir, f)))
        .filter((p): p is Project => p !== null);
}

/** 获取所有已发布文章，按发布时间倒序。 */
export function getPublishedPosts(): Post[] {
    const dir = path.join(DATA_DIR, 'posts');
    return safeReadDir(dir)
        .filter(f => f.endsWith('.json'))
        .map(f => safeReadJSON<Post>(path.join(dir, f)))
        .filter((p): p is Post => p !== null && p.published_at != null)
        .sort((a, b) => (b.published_at || '').localeCompare(a.published_at || ''));
}

/** 获取所有文章（含草稿），用于 getStaticPaths 枚举。 */
export function getAllPosts(): Post[] {
    const dir = path.join(DATA_DIR, 'posts');
    return safeReadDir(dir)
        .filter(f => f.endsWith('.json'))
        .map(f => safeReadJSON<Post>(path.join(dir, f)))
        .filter((p): p is Post => p !== null);
}

/** 获取指定 slug 的文章。 */
export function getPost(slug: string): Post | null {
    return safeReadJSON<Post>(path.join(DATA_DIR, 'posts', `${slug}.json`));
}

/** 获取所有分类定义（data/categories.json）。 */
export function getCategories(): Category[] {
    const data = safeReadJSON<Category[]>(path.join(DATA_DIR, 'categories.json'));
    return data || [];
}

/** 获取指定项目的快照数据，可限制最近 N 天。 */
export function getProjectSnapshots(projectId: string, days?: number): Snapshot[] {
    const dir = path.join(DATA_DIR, 'snapshots');
    const files = safeReadDir(dir).filter(f => f.endsWith('.jsonl')).sort().reverse();

    const targetFiles = days ? files.slice(0, days) : files;
    const snapshots: Snapshot[] = [];

    for (const file of targetFiles) {
        try {
            const content = fs.readFileSync(path.join(dir, file), 'utf-8');
            for (const line of content.split('\n')) {
                if (!line.trim()) continue;
                const snap = JSON.parse(line) as Snapshot;
                if (snap.project_id === projectId) {
                    snapshots.push(snap);
                }
            }
        } catch {
            // 跳过损坏的 JSONL 文件
        }
    }

    return snapshots.sort((a, b) => a.date.localeCompare(b.date));
}
