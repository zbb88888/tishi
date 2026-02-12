/* empty css                                  */
import { g as createComponent, j as renderComponent, r as renderTemplate, m as maybeRenderHead } from '../chunks/astro/server_CCdsDrdo.mjs';
import 'kleur/colors';
import { $ as $$BaseLayout } from '../chunks/BaseLayout_CnYfUBOf.mjs';
import { b as getCategories } from '../chunks/api_CbtYqGEw.mjs';
export { renderers } from '../renderers.mjs';

const $$Categories = createComponent(async ($$result, $$props, $$slots) => {
  let categories = [];
  let error = null;
  try {
    const res = await getCategories();
    categories = res.data;
  } catch (e) {
    error = "\u6682\u65F6\u65E0\u6CD5\u52A0\u8F7D\u5206\u7C7B\u6570\u636E";
  }
  return renderTemplate`${renderComponent($$result, "BaseLayout", $$BaseLayout, { "title": "\u5206\u7C7B" }, { "default": async ($$result2) => renderTemplate` ${maybeRenderHead()}<div class="mb-8"> <h1 class="text-3xl font-bold text-gray-900">📂 项目分类</h1> <p class="mt-2 text-gray-600">按 AI 子领域浏览热门开源项目</p> </div> ${error ? renderTemplate`<div class="card text-center text-gray-500 py-12"> <p>${error}</p> </div>` : renderTemplate`<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"> ${categories.map((cat) => renderTemplate`<div class="card"> <h3 class="text-lg font-semibold text-gray-900">${cat.name}</h3> ${cat.description && renderTemplate`<p class="mt-1 text-sm text-gray-500">${cat.description}</p>`} <div class="mt-4 flex items-center justify-between"> <span class="badge-blue">${cat.project_count} 个项目</span> <span class="text-xs text-gray-400">${cat.slug}</span> </div> </div>`)} </div>`}` })}`;
}, "/root/g/tishi/web/src/pages/categories.astro", void 0);

const $$file = "/root/g/tishi/web/src/pages/categories.astro";
const $$url = "/categories";

const _page = /*#__PURE__*/Object.freeze(/*#__PURE__*/Object.defineProperty({
  __proto__: null,
  default: $$Categories,
  file: $$file,
  url: $$url
}, Symbol.toStringTag, { value: 'Module' }));

const page = () => _page;

export { page };
