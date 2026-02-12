const API_BASE = import.meta.env.PUBLIC_API_URL || '';

export interface Project {
    id: number;
    full_name: string;
    description: string | null;
    language: string | null;
    license: string | null;
    stars: number;
    forks: number;
    open_issues: number;
    score: number;
    rank: number | null;
    topics: string[];
    pushed_at: string | null;
    github_url?: string;
}

export interface BlogPost {
    id: number;
    title: string;
    slug: string;
    type: string;
    published_at: string | null;
    excerpt?: string;
    content?: string;
}

export interface Category {
    id: number;
    name: string;
    slug: string;
    description: string | null;
    project_count: number;
}

export interface TrendPoint {
    date: string;
    stars: number;
    forks: number;
    open_issues: number;
    score: number | null;
    rank: number | null;
}

export interface ApiResponse<T> {
    data: T;
    meta?: {
        total: number;
        page: number;
        per_page: number;
        total_pages: number;
    };
}

async function fetchApi<T>(path: string): Promise<ApiResponse<T>> {
    const res = await fetch(`${API_BASE}${path}`);
    if (!res.ok) {
        throw new Error(`API error: ${res.status} ${res.statusText}`);
    }
    return res.json();
}

export async function getRankings(page = 1, perPage = 20) {
    return fetchApi<Project[]>(`/api/v1/rankings?page=${page}&per_page=${perPage}`);
}

export async function getProjects(page = 1, perPage = 20) {
    return fetchApi<Project[]>(`/api/v1/projects?page=${page}&per_page=${perPage}`);
}

export async function getProject(id: number) {
    return fetchApi<Project>(`/api/v1/projects/${id}`);
}

export async function getProjectTrends(id: number, days = 30) {
    return fetchApi<TrendPoint[]>(`/api/v1/projects/${id}/trends?days=${days}`);
}

export async function getPosts(page = 1, perPage = 10) {
    return fetchApi<BlogPost[]>(`/api/v1/posts?page=${page}&per_page=${perPage}`);
}

export async function getPost(slug: string) {
    return fetchApi<BlogPost>(`/api/v1/posts/${slug}`);
}

export async function getCategories() {
    return fetchApi<Category[]>(`/api/v1/categories`);
}
