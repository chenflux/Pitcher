import axios from "axios";

const api = axios.create({ baseURL: "/api" });

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      const url = err.config?.url || "";
      if (!url.includes("/auth/login")) {
        localStorage.removeItem("token");
        window.location.href = "/login";
      }
    }
    return Promise.reject(err);
  }
);

export default api;

export const authApi = {
  getCaptcha: () => api.get("/auth/captcha"),
  login: (username: string, password: string, captcha?: string, captchaId?: string) =>
    api.post("/auth/login", { username, password, captcha, captcha_id: captchaId }),
  getProfile: () => api.get("/auth/profile"),
  changePassword: (old_password: string, new_password: string) =>
    api.put("/auth/password", { old_password, new_password }),
};

export const nodeApi = {
  list: () => api.get("/nodes"),
  get: (id: string) => api.get(`/nodes/${id}`),
  register: (data: any) => api.post("/nodes/register", data),
  heartbeat: (id: string, data: any) => api.post(`/nodes/${id}/heartbeat`, data),
  delete: (id: string) => api.delete(`/nodes/${id}`),
};

export const serviceApi = {
  list: (nodeId?: string) =>
    api.get("/services", { params: nodeId ? { node_id: nodeId } : {} }),
  get: (id: number) => api.get(`/services/${id}`),
  create: (data: any) => api.post("/services", data),
  update: (id: number, data: any) => api.put(`/services/${id}`, data),
  delete: (id: number) => api.delete(`/services/${id}`),
  start: (id: number) => api.post(`/services/${id}/start`),
  stop: (id: number) => api.post(`/services/${id}/stop`),
  restart: (id: number) => api.post(`/services/${id}/restart`),
};

export const logApi = {
  query: (params: any) => api.get("/logs", { params }),
  stats: (hours?: number) => api.get("/logs/stats", { params: { hours } }),
  trend: (honeypot_type?: number, range?: string) =>
    api.get("/logs/trend", { params: { honeypot_type, range } }),
  export: (params: any) => api.get("/logs/export", { params, responseType: "blob" }),
};

export const configApi = {
  list: () => api.get("/configs"),
  get: (id: number) => api.get(`/configs/${id}`),
  create: (data: any) => api.post("/configs", data),
  update: (id: number, data: any) => api.put(`/configs/${id}`, data),
  delete: (id: number) => api.delete(`/configs/${id}`),
  push: (id: number) => api.post(`/configs/${id}/push`),
};

export const ruleApi = {
  listGroups: (protocol?: string) =>
    api.get("/rules/groups", { params: protocol ? { protocol } : {} }),
  getGroup: (id: number) => api.get(`/rules/groups/${id}`),
  createGroup: (data: any) => api.post("/rules/groups", data),
  updateGroup: (id: number, data: any) => api.put(`/rules/groups/${id}`, data),
  deleteGroup: (id: number) => api.delete(`/rules/groups/${id}`),
  createRule: (groupId: number, data: any) =>
    api.post(`/rules/groups/${groupId}/rules`, data),
  updateRule: (groupId: number, ruleId: number, data: any) =>
    api.put(`/rules/groups/${groupId}/rules/${ruleId}`, data),
  deleteRule: (groupId: number, ruleId: number) =>
    api.delete(`/rules/groups/${groupId}/rules/${ruleId}`),
};

export const settingsApi = {
  getAgentToken: () => api.get("/settings/agent-token"),
  regenerateAgentToken: () => api.put("/settings/agent-token"),
  getVersion: () => api.get("/version"),
  getAgentDownloads: () => api.get("/downloads/agents"),
};
