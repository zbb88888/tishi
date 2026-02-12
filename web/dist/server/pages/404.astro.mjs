/* empty css                                  */
import { g as createComponent, j as renderComponent, r as renderTemplate, m as maybeRenderHead } from '../chunks/astro/server_CCdsDrdo.mjs';
import 'kleur/colors';
import { $ as $$BaseLayout } from '../chunks/BaseLayout_CnYfUBOf.mjs';
export { renderers } from '../renderers.mjs';

const $$404 = createComponent(($$result, $$props, $$slots) => {
  return renderTemplate`${renderComponent($$result, "BaseLayout", $$BaseLayout, { "title": "\u9875\u9762\u672A\u627E\u5230" }, { "default": ($$result2) => renderTemplate` ${maybeRenderHead()}<div class="flex flex-col items-center justify-center py-24 text-center"> <div class="text-8xl font-bold text-gray-200">404</div> <h1 class="mt-4 text-2xl font-semibold text-gray-900">页面未找到</h1> <p class="mt-2 text-gray-500">您访问的页面不存在或已被移除</p> <div class="mt-8 flex gap-4"> <a href="/" class="inline-flex items-center gap-2 rounded-lg bg-primary-600 px-5 py-2.5 text-sm font-medium text-white hover:bg-primary-700 transition-colors">
← 返回排行榜
</a> <a href="/blog" class="inline-flex items-center gap-2 rounded-lg border border-gray-300 px-5 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors">
浏览博客
</a> </div> </div> ` })}`;
}, "/root/g/tishi/web/src/pages/404.astro", void 0);

const $$file = "/root/g/tishi/web/src/pages/404.astro";
const $$url = "/404";

const _page = /*#__PURE__*/Object.freeze(/*#__PURE__*/Object.defineProperty({
  __proto__: null,
  default: $$404,
  file: $$file,
  url: $$url
}, Symbol.toStringTag, { value: 'Module' }));

const page = () => _page;

export { page };
