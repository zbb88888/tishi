/* empty css                                  */
import { f as createAstro, g as createComponent, m as maybeRenderHead, i as addAttribute, r as renderTemplate, j as renderComponent } from '../chunks/astro/server_CCdsDrdo.mjs';
import 'kleur/colors';
import { $ as $$BaseLayout } from '../chunks/BaseLayout_CnYfUBOf.mjs';
import 'clsx';
import { a as getPosts } from '../chunks/api_CbtYqGEw.mjs';
export { renderers } from '../renderers.mjs';

const $$Astro$1 = createAstro("https://tishi.dev");
const $$Pagination = createComponent(($$result, $$props, $$slots) => {
  const Astro2 = $$result.createAstro($$Astro$1, $$props, $$slots);
  Astro2.self = $$Pagination;
  const { currentPage, totalPages, baseUrl } = Astro2.props;
  function getPageNumbers(current, total) {
    if (total <= 7) {
      return Array.from({ length: total }, (_, i) => i + 1);
    }
    const pages2 = [1];
    if (current > 3) pages2.push("...");
    const start = Math.max(2, current - 1);
    const end = Math.min(total - 1, current + 1);
    for (let i = start; i <= end; i++) {
      pages2.push(i);
    }
    if (current < total - 2) pages2.push("...");
    pages2.push(total);
    return pages2;
  }
  const pages = getPageNumbers(currentPage, totalPages);
  function pageUrl(page) {
    const sep = baseUrl.includes("?") ? "&" : "?";
    return `${baseUrl}${sep}page=${page}`;
  }
  return renderTemplate`${totalPages > 1 && renderTemplate`${maybeRenderHead()}<nav class="mt-8 flex items-center justify-center gap-1" aria-label="分页导航">${currentPage > 1 ? renderTemplate`<a${addAttribute(pageUrl(currentPage - 1), "href")} class="rounded-lg px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 transition-colors">
← 上一页
</a>` : renderTemplate`<span class="rounded-lg px-3 py-2 text-sm text-gray-300 cursor-not-allowed">
← 上一页
</span>`}${pages.map(
    (p) => p === "..." ? renderTemplate`<span class="px-2 py-2 text-sm text-gray-400">…</span>` : p === currentPage ? renderTemplate`<span class="rounded-lg bg-primary-600 px-3.5 py-2 text-sm font-medium text-white">${p}</span>` : renderTemplate`<a${addAttribute(pageUrl(p), "href")} class="rounded-lg px-3.5 py-2 text-sm text-gray-600 hover:bg-gray-100 transition-colors">${p}</a>`
  )}${currentPage < totalPages ? renderTemplate`<a${addAttribute(pageUrl(currentPage + 1), "href")} class="rounded-lg px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 transition-colors">
下一页 →
</a>` : renderTemplate`<span class="rounded-lg px-3 py-2 text-sm text-gray-300 cursor-not-allowed">
下一页 →
</span>`}</nav>`}`;
}, "/root/g/tishi/web/src/components/Pagination.astro", void 0);

const $$Astro = createAstro("https://tishi.dev");
const $$Index = createComponent(async ($$result, $$props, $$slots) => {
  const Astro2 = $$result.createAstro($$Astro, $$props, $$slots);
  Astro2.self = $$Index;
  const currentPage = Number(Astro2.url.searchParams.get("page") || "1");
  const perPage = 10;
  let posts = [];
  let totalPages = 1;
  let error = null;
  try {
    const res = await getPosts(currentPage, perPage);
    posts = res.data;
    totalPages = res.meta?.total_pages || 1;
  } catch (e) {
    error = "\u6682\u65F6\u65E0\u6CD5\u52A0\u8F7D\u535A\u5BA2\u6587\u7AE0";
  }
  function formatDate(dateStr) {
    if (!dateStr) return "";
    return new Date(dateStr).toLocaleDateString("zh-CN", {
      year: "numeric",
      month: "long",
      day: "numeric"
    });
  }
  function typeLabel(type) {
    const labels = {
      weekly: "\u{1F4CA} \u5468\u62A5",
      monthly: "\u{1F4C8} \u6708\u62A5"
    };
    return labels[type] || type;
  }
  return renderTemplate`${renderComponent($$result, "BaseLayout", $$BaseLayout, { "title": "\u535A\u5BA2" }, { "default": async ($$result2) => renderTemplate` ${maybeRenderHead()}<div class="mb-8"> <h1 class="text-3xl font-bold text-gray-900">📝 博客</h1> <p class="mt-2 text-gray-600">AI 开源趋势周报、月报与深度分析</p> </div> ${error ? renderTemplate`<div class="card text-center text-gray-500 py-12"> <p>${error}</p> </div>` : posts.length === 0 ? renderTemplate`<div class="card text-center text-gray-500 py-12"> <p>暂无文章</p> <p class="mt-2 text-sm">运行 <code class="bg-gray-100 px-2 py-1 rounded">tishi generate weekly</code> 生成首篇周报</p> </div>` : renderTemplate`<div class="space-y-4"> ${posts.map((post) => renderTemplate`<a${addAttribute(`/blog/${post.slug}`, "href")} class="card block group"> <div class="flex items-start justify-between"> <div> <span class="text-xs text-gray-400">${typeLabel(post.type)}</span> <h2 class="mt-1 text-lg font-semibold text-gray-900 group-hover:text-primary-600 transition-colors"> ${post.title} </h2> ${post.excerpt && renderTemplate`<p class="mt-2 text-sm text-gray-500 line-clamp-2">${post.excerpt}</p>`} </div> <time class="shrink-0 text-xs text-gray-400"> ${formatDate(post.published_at)} </time> </div> </a>`)} </div>

    ${renderComponent($$result2, "Pagination", $$Pagination, { "currentPage": currentPage, "totalPages": totalPages, "baseUrl": "/blog" })}`}` })}`;
}, "/root/g/tishi/web/src/pages/blog/index.astro", void 0);

const $$file = "/root/g/tishi/web/src/pages/blog/index.astro";
const $$url = "/blog";

const _page = /*#__PURE__*/Object.freeze(/*#__PURE__*/Object.defineProperty({
  __proto__: null,
  default: $$Index,
  file: $$file,
  url: $$url
}, Symbol.toStringTag, { value: 'Module' }));

const page = () => _page;

export { page };
