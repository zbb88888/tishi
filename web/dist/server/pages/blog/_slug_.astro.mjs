/* empty css                                     */
import { f as createAstro, g as createComponent, j as renderComponent, r as renderTemplate, m as maybeRenderHead, u as unescapeHTML } from '../../chunks/astro/server_CCdsDrdo.mjs';
import 'kleur/colors';
import { g as getPost, $ as $$BaseLayout } from '../../chunks/api_CaJSCn5g.mjs';
export { renderers } from '../../renderers.mjs';

const $$Astro = createAstro("https://tishi.dev");
const $$slug = createComponent(async ($$result, $$props, $$slots) => {
  const Astro2 = $$result.createAstro($$Astro, $$props, $$slots);
  Astro2.self = $$slug;
  const { slug } = Astro2.params;
  let post = null;
  let error = null;
  try {
    if (slug) {
      const res = await getPost(slug);
      post = res.data;
    }
  } catch (e) {
    error = "\u6587\u7AE0\u672A\u627E\u5230";
  }
  return renderTemplate`${renderComponent($$result, "BaseLayout", $$BaseLayout, { "title": post?.title || "\u6587\u7AE0" }, { "default": async ($$result2) => renderTemplate`${error || !post ? renderTemplate`${maybeRenderHead()}<div class="card text-center text-gray-500 py-12"> <p class="text-lg">${error || "\u6587\u7AE0\u4E0D\u5B58\u5728"}</p> <a href="/blog" class="mt-4 inline-block text-primary-600 hover:underline">← 返回博客列表</a> </div>` : renderTemplate`<article class="prose prose-gray mx-auto max-w-3xl"> <a href="/blog" class="text-sm text-gray-500 hover:text-primary-600 no-underline">← 返回博客列表</a> <h1 class="mt-4">${post.title}</h1> ${post.published_at && renderTemplate`<time class="text-sm text-gray-400"> ${new Date(post.published_at).toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "long",
    day: "numeric"
  })} </time>`} <div class="mt-8 whitespace-pre-wrap">${unescapeHTML(post.content)}</div> </article>`}` })}`;
}, "/root/g/tishi/web/src/pages/blog/[slug].astro", void 0);

const $$file = "/root/g/tishi/web/src/pages/blog/[slug].astro";
const $$url = "/blog/[slug]";

const _page = /*#__PURE__*/Object.freeze(/*#__PURE__*/Object.defineProperty({
  __proto__: null,
  default: $$slug,
  file: $$file,
  url: $$url
}, Symbol.toStringTag, { value: 'Module' }));

const page = () => _page;

export { page };
