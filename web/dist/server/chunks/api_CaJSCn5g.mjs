import { f as createAstro, g as createComponent, i as addAttribute, l as renderHead, n as renderSlot, r as renderTemplate } from './astro/server_CCdsDrdo.mjs';
import 'kleur/colors';
import 'clsx';
/* empty css                          */

const $$Astro = createAstro("https://tishi.dev");
const $$BaseLayout = createComponent(($$result, $$props, $$slots) => {
  const Astro2 = $$result.createAstro($$Astro, $$props, $$slots);
  Astro2.self = $$BaseLayout;
  const { title, description = "tishi \u2014 \u8FFD\u8E2A GitHub AI Top 100 \u70ED\u95E8\u5F00\u6E90\u9879\u76EE" } = Astro2.props;
  return renderTemplate`<html lang="zh-CN"> <head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><meta name="description"${addAttribute(description, "content")}><link rel="icon" type="image/svg+xml" href="/favicon.svg"><title>${title} | tishi</title>${renderHead()}</head> <body class="min-h-screen bg-gray-50 text-gray-900 antialiased"> <header class="sticky top-0 z-50 border-b border-gray-200 bg-white/80 backdrop-blur-md"> <nav class="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6"> <a href="/" class="flex items-center gap-2 text-xl font-bold text-primary-600"> <span>📡</span> <span>tishi</span> </a> <ul class="flex items-center gap-6 text-sm font-medium text-gray-600"> <li><a href="/" class="hover:text-primary-600 transition-colors">排行榜</a></li> <li><a href="/categories" class="hover:text-primary-600 transition-colors">分类</a></li> <li><a href="/blog" class="hover:text-primary-600 transition-colors">博客</a></li> <li> <a href="https://github.com/zbb88888/tishi" target="_blank" rel="noopener" class="hover:text-primary-600 transition-colors">GitHub</a> </li> </ul> </nav> </header> <main class="mx-auto max-w-7xl px-4 py-8 sm:px-6"> ${renderSlot($$result, $$slots["default"])} </main> <footer class="border-t border-gray-200 bg-white py-8"> <div class="mx-auto max-w-7xl px-4 text-center text-sm text-gray-500 sm:px-6"> <p>tishi — 自动追踪 GitHub AI Top 100 热门项目</p> <p class="mt-1">数据每日自动更新 · 开源项目</p> </div> </footer> </body></html>`;
}, "/root/g/tishi/web/src/layouts/BaseLayout.astro", void 0);

const API_BASE = "";
async function fetchApi(path) {
  const res = await fetch(`${API_BASE}${path}`);
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}
async function getRankings(page = 1, perPage = 20) {
  return fetchApi(`/api/v1/rankings?page=${page}&per_page=${perPage}`);
}
async function getProject(id) {
  return fetchApi(`/api/v1/projects/${id}`);
}
async function getPosts(page = 1, perPage = 10) {
  return fetchApi(`/api/v1/posts?page=${page}&per_page=${perPage}`);
}
async function getPost(slug) {
  return fetchApi(`/api/v1/posts/${slug}`);
}
async function getCategories() {
  return fetchApi(`/api/v1/categories`);
}

export { $$BaseLayout as $, getPosts as a, getCategories as b, getProject as c, getRankings as d, getPost as g };
