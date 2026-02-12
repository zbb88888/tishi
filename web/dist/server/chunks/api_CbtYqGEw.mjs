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

export { getPosts as a, getCategories as b, getProject as c, getRankings as d, getPost as g };
