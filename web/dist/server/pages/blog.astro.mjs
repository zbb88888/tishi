/* empty css                                  */
import { g as createComponent, j as renderComponent, r as renderTemplate, m as maybeRenderHead, i as addAttribute } from '../chunks/astro/server_CCdsDrdo.mjs';
import 'kleur/colors';
import { a as getPosts, $ as $$BaseLayout } from '../chunks/api_CaJSCn5g.mjs';
export { renderers } from '../renderers.mjs';

const $$Index = createComponent(async ($$result, $$props, $$slots) => {
  let posts = [];
  let error = null;
  try {
    const res = await getPosts(1, 20);
    posts = res.data;
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
  return renderTemplate`${renderComponent($$result, "BaseLayout", $$BaseLayout, { "title": "\u535A\u5BA2" }, { "default": async ($$result2) => renderTemplate` ${maybeRenderHead()}<div class="mb-8"> <h1 class="text-3xl font-bold text-gray-900">📝 博客</h1> <p class="mt-2 text-gray-600">AI 开源趋势周报、月报与深度分析</p> </div> ${error ? renderTemplate`<div class="card text-center text-gray-500 py-12"> <p>${error}</p> </div>` : posts.length === 0 ? renderTemplate`<div class="card text-center text-gray-500 py-12"> <p>暂无文章</p> <p class="mt-2 text-sm">运行 <code class="bg-gray-100 px-2 py-1 rounded">tishi generate weekly</code> 生成首篇周报</p> </div>` : renderTemplate`<div class="space-y-4"> ${posts.map((post) => renderTemplate`<a${addAttribute(`/blog/${post.slug}`, "href")} class="card block group"> <div class="flex items-start justify-between"> <div> <span class="text-xs text-gray-400">${typeLabel(post.type)}</span> <h2 class="mt-1 text-lg font-semibold text-gray-900 group-hover:text-primary-600 transition-colors"> ${post.title} </h2> ${post.excerpt && renderTemplate`<p class="mt-2 text-sm text-gray-500 line-clamp-2">${post.excerpt}</p>`} </div> <time class="shrink-0 text-xs text-gray-400"> ${formatDate(post.published_at)} </time> </div> </a>`)} </div>`}` })}`;
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
