import axios from "axios";
import { useAuthStore } from "@/store/auth";
import { showToast } from "@/lib/toast";

const client = axios.create({
  baseURL: `${import.meta.env.VITE_BASE_PATH ?? ""}/v1`,
});

// Request interceptor: attach Authorization header
client.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: unwrap data and handle errors
client.interceptors.response.use(
  (response) => {
    const body = response.data;
    if (body && body.code === 0) {
      return body.data;
    }
    return response.data;
  },
  (error) => {
    if (!error.response) {
      return Promise.reject(error);
    }

    const { status } = error.response;

    switch (status) {
      case 401: {
        const basePath = import.meta.env.VITE_BASE_PATH ?? "";
        useAuthStore.getState().clearAuth();
        if (window.location.pathname !== `${basePath}/login`) {
          window.location.href = `${basePath}/login`;
        }
        break;
      }
      case 403:
        showToast("权限不足", "error");
        useAuthStore.getState().fetchPermissions();
        break;
      case 404:
        showToast("资源不存在", "error");
        break;
      case 422:
        // Re-throw for component-level handling
        break;
      case 500:
        showToast("服务器错误，请稍后重试", "error");
        break;
    }

    return Promise.reject(error);
  },
);

export default client;
