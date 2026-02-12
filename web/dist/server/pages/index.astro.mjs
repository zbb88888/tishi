/* empty css                                  */
import { g as createComponent, j as renderComponent, r as renderTemplate, m as maybeRenderHead, i as addAttribute } from '../chunks/astro/server_CCdsDrdo.mjs';
import 'kleur/colors';
import { d as getRankings, $ as $$BaseLayout } from '../chunks/api_CaJSCn5g.mjs';
export { renderers } from '../renderers.mjs';

const $$Index = createComponent(async ($$result, $$props, $$slots) => {
  let projects = [];
  let error = null;
  try {
    const res = await getRankings(1, 100);
    projects = res.data;
  } catch (e) {
    error = "\u6682\u65F6\u65E0\u6CD5\u52A0\u8F7D\u6392\u884C\u699C\u6570\u636E\uFF0C\u8BF7\u786E\u8BA4\u540E\u7AEF\u670D\u52A1\u5DF2\u542F\u52A8\u3002";
  }
  function languageColor(lang) {
    const colors = {
      Python: "bg-blue-500",
      TypeScript: "bg-blue-600",
      JavaScript: "bg-yellow-400",
      Go: "bg-cyan-500",
      Rust: "bg-orange-600",
      "C++": "bg-pink-600",
      Java: "bg-red-500",
      Jupyter: "bg-orange-400"
    };
    return colors[lang || ""] || "bg-gray-400";
  }
  function formatNumber(n) {
    if (n >= 1e3) return `${(n / 1e3).toFixed(1)}k`;
    return String(n);
  }
  return renderTemplate`${renderComponent($$result, "BaseLayout", $$BaseLayout, { "title": "AI \u5F00\u6E90\u6392\u884C\u699C" }, { "default": async ($$result2) => renderTemplate` ${maybeRenderHead()}<div class="mb-8"> <h1 class="text-3xl font-bold text-gray-900">🔥 AI 开源项目排行榜</h1> <p class="mt-2 text-gray-600">GitHub AI 相关 Top 100 热门项目，数据每日自动更新</p> </div> ${error ? renderTemplate`<div class="card text-center text-gray-500 py-12"> <p class="text-lg">${error}</p> <p class="mt-2 text-sm">运行 <code class="bg-gray-100 px-2 py-1 rounded">make run</code> 启动后端服务</p> </div>` : projects.length === 0 ? renderTemplate`<div class="card text-center text-gray-500 py-12"> <p class="text-lg">暂无数据</p> <p class="mt-2 text-sm">运行 <code class="bg-gray-100 px-2 py-1 rounded">tishi collect</code> 开始首次数据采集</p> </div>` : renderTemplate`<div class="overflow-x-auto"> <table class="w-full text-left text-sm"> <thead class="bg-gray-100 text-xs uppercase text-gray-600"> <tr> <th class="px-4 py-3 w-16">排名</th> <th class="px-4 py-3">项目</th> <th class="px-4 py-3 w-24">语言</th> <th class="px-4 py-3 w-24 text-right">⭐ Star</th> <th class="px-4 py-3 w-24 text-right">🍴 Fork</th> <th class="px-4 py-3 w-20 text-right">评分</th> </tr> </thead> <tbody class="divide-y divide-gray-200"> ${projects.map((p) => renderTemplate`<tr class="hover:bg-gray-50 transition-colors"> <td class="px-4 py-3 font-bold text-gray-400"> ${p.rank != null ? renderTemplate`<span${addAttribute(p.rank <= 3 ? "text-primary-600 text-lg" : "", "class")}>
#${p.rank} </span>` : "-"} </td> <td class="px-4 py-3"> <div> <a${addAttribute(`/projects/${p.id}`, "href")} class="font-medium text-gray-900 hover:text-primary-600 transition-colors"> ${p.full_name} </a> ${p.description && renderTemplate`<p class="mt-0.5 text-xs text-gray-500 line-clamp-1">${p.description}</p>`} </div> </td> <td class="px-4 py-3"> ${p.language && renderTemplate`<span class="inline-flex items-center gap-1.5 text-xs"> <span${addAttribute(`h-2.5 w-2.5 rounded-full ${languageColor(p.language)}`, "class")}></span> ${p.language} </span>`} </td> <td class="px-4 py-3 text-right font-mono text-xs"> ${formatNumber(p.stars)} </td> <td class="px-4 py-3 text-right font-mono text-xs text-gray-500"> ${formatNumber(p.forks)} </td> <td class="px-4 py-3 text-right"> <span class="badge-blue">${p.score.toFixed(1)}</span> </td> </tr>`)} </tbody> </table> </div>`}` })}`;
}, "/root/g/tishi/web/src/pages/index.astro", void 0);

const $$file = "/root/g/tishi/web/src/pages/index.astro";
const $$url = "";

const _page = /*#__PURE__*/Object.freeze(/*#__PURE__*/Object.defineProperty({
  __proto__: null,
  default: $$Index,
  file: $$file,
  url: $$url
}, Symbol.toStringTag, { value: 'Module' }));

const page = () => _page;

export { page };
