const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8082/api/v1";

interface ApiResponse<T> {
  code: number;
  msg: string;
  data: T;
}

class ApiClient {
  private baseUrl: string;
  private token: string | null = null;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
    if (typeof window !== "undefined") {
      this.token = localStorage.getItem("token") || sessionStorage.getItem("token");
    }
  }

  setToken(token: string) {
    this.token = token;
  }

  clearToken() {
    this.token = null;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    // Always read latest token from storage to avoid stale token issue
    if (typeof window !== "undefined") {
      this.token = localStorage.getItem("token") || sessionStorage.getItem("token") || this.token;
    }

    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...options.headers,
    };

    if (this.token) {
      (headers as Record<string, string>)["Authorization"] = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        ...options,
        headers,
      });

      const data = await response.json();

      if (data.code !== 10000) {
        throw new Error(data.msg || "请求失败");
      }

      return data;
    } catch (error) {
      // Network error or API unavailable - rethrow for caller to handle
      throw error;
    }
  }

  // Auth endpoints
  async signup(username: string, password: string, rePassword: string) {
    return this.request("/signup", {
      method: "POST",
      body: JSON.stringify({ username, password, re_password: rePassword }),
    });
  }

  async login(username: string, password: string) {
    const response = await this.request<{ user_id: string; user_name: string; token: string }>(
      "/login",
      {
        method: "POST",
        body: JSON.stringify({ username, password }),
      }
    );
    
    if (response.data.token) {
      this.setToken(response.data.token);
    }
    
    return response;
  }

  // Community endpoints
  async getCommunities() {
    return this.request<Array<{ community_id: string; name: string }>>(
      "/community"
    );
  }

  async getCommunityDetail(id: string) {
    return this.request<{
      community_id: string;
      name: string;
      introduction: string;
      create_time: string;
    }>(`/community/${id}`);
  }

  // Post endpoints
  async getPosts(params?: {
    page?: number;
    size?: number;
    community_id?: string;
    sort_by?: string;
    order?: string;
  }) {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.size) searchParams.set("size", params.size.toString());
    if (params?.community_id) searchParams.set("community_id", params.community_id);
    if (params?.sort_by) searchParams.set("sort_by", params.sort_by);
    if (params?.order) searchParams.set("order", params.order);

    const query = searchParams.toString();
    return this.request<{
      list: Array<{
        post_id: string;
        title: string;
        summary: string;
        user_name: string;
        community_name: string;
        create_time: string;
        like_count: number;
        comment_count: number;
      }>;
      total: number;
    }>(`/posts${query ? `?${query}` : ""}`);
  }

  async getPostDetail(id: string) {
    return this.request<{
      post_id: string;
      title: string;
      content: string;
      author_name: string;
      community: {
        community_id: string;
        name: string;
        introduction: string;
      };
      create_time: string;
      like_count: number;
      comment_count: number;
    }>(`/post_detail/${id}`);
  }

  async createPost(title: string, content: string, communityId: string) {
    return this.request("/create_post", {
      method: "POST",
      body: JSON.stringify({
        title,
        content,
        community_id: communityId,
      }),
    });
  }

  // Vote endpoints
  async vote(postId: string, direction: number) {
    return this.request("/vote", {
      method: "POST",
      body: JSON.stringify({
        post_id: postId,
        direction,
      }),
    });
  }
}

export const apiClient = new ApiClient(API_BASE_URL);
export type { ApiResponse };
