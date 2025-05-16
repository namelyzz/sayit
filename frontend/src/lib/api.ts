const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8082/api/v1";

interface ApiResponse<T> {
  code: number;
  msg: string;
  data: T;
}

export interface UserSummary {
  user_id: string;
  user_name: string;
}

export interface UserProfile {
  user_id: string;
  user_name: string;
  signature: string;
  create_time: string;
  post_count: number;
  post_score: number;
  follower_count: number;
  following_count: number;
  is_following: boolean;
  is_self: boolean;
}

export interface UserFollowItem {
  user_id: string;
  user_name: string;
  signature: string;
  is_following: boolean;
  is_followed_by: boolean;
  is_mutual: boolean;
}

export interface UserFollowListResponse {
  list: UserFollowItem[];
  total: number;
}

export interface CommunitySummary {
  community_id: string;
  name: string;
}

export interface HotCommunity {
  community_id: string;
  community_name: string;
}

export interface CommunityDetail {
  community_id: string;
  name: string;
  introduction: string;
  create_time: string;
}

export interface PostListItem {
  post_id: string;
  title: string;
  summary: string;
  author_id: string;
  user_name: string;
  community_id: string;
  community_name: string;
  create_time: string;
  like_count: number;
  vote_count: number;
  comment_count: number;
  current_user_vote: number;
}

export interface PostsResponse {
  list: PostListItem[];
  total: number;
}

export interface PostDetail {
  post_id: string;
  title: string;
  content: string;
  author_id: string;
  author_name: string;
  community: {
    community_id: string;
    name: string;
    introduction: string;
  };
  create_time: string;
  like_count: number;
  vote_count: number;
  comment_count: number;
  current_user_vote: number;
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

  private clearAuthStorage() {
    if (typeof window === "undefined") return;

    localStorage.removeItem("token");
    localStorage.removeItem("user_id");
    localStorage.removeItem("user_name");
    sessionStorage.removeItem("token");
    sessionStorage.removeItem("user_id");
    sessionStorage.removeItem("user_name");
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
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

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers,
    });

    const data = (await response.json()) as ApiResponse<T>;

    if (data.code !== 10000) {
      if (data.code === 10012 || data.msg.includes("token")) {
        this.clearAuthStorage();
        this.clearToken();
      }
      throw new Error(data.msg || "请求失败，请稍后再试");
    }

    return data;
  }

  async signup(username: string, password: string, rePassword: string) {
    return this.request("/signup", {
      method: "POST",
      body: JSON.stringify({ username, password, re_password: rePassword }),
    });
  }

  async login(username: string, password: string) {
    const response = await this.request<UserSummary & { token: string }>("/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });

    if (response.data.token) {
      this.setToken(response.data.token);
    }

    return response;
  }

  async getUserProfile(id: string) {
    return this.request<UserProfile>(`/users/${id}`);
  }

  async getMe() {
    return this.request<UserProfile>("/me");
  }

  async updateMe(signature: string) {
    return this.request<UserProfile>("/me", {
      method: "PATCH",
      body: JSON.stringify({ signature }),
    });
  }

  async followUser(id: string) {
    return this.request<{ is_following: boolean }>(`/users/${id}/follow`, {
      method: "POST",
    });
  }

  async unfollowUser(id: string) {
    return this.request<{ is_following: boolean }>(`/users/${id}/follow`, {
      method: "DELETE",
    });
  }

  async getUserFollowers(id: string, params?: { page?: number; size?: number }) {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.size) searchParams.set("size", params.size.toString());
    const query = searchParams.toString();
    return this.request<UserFollowListResponse>(`/users/${id}/followers${query ? `?${query}` : ""}`);
  }

  async getUserFollowing(id: string, params?: { page?: number; size?: number }) {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.size) searchParams.set("size", params.size.toString());
    const query = searchParams.toString();
    return this.request<UserFollowListResponse>(`/users/${id}/following${query ? `?${query}` : ""}`);
  }

  async getCommunities() {
    return this.request<CommunitySummary[]>("/community");
  }

  async getHotCommunities(limit?: number) {
    const query = limit ? `?limit=${limit}` : "";
    return this.request<HotCommunity[]>(`/hot_communities${query}`);
  }

  async getRandomCommunities(limit?: number) {
    const query = limit ? `?limit=${limit}` : "";
    return this.request<CommunitySummary[]>(`/random_communities${query}`);
  }

  async followCommunity(communityId: string) {
    return this.request("/follow", {
      method: "POST",
      body: JSON.stringify({ community_id: communityId }),
    });
  }

  async unfollowCommunity(communityId: string) {
    return this.request("/unfollow", {
      method: "POST",
      body: JSON.stringify({ community_id: communityId }),
    });
  }

  async isFollowedCommunity(communityId: string) {
    return this.request<{ is_followed: boolean }>(`/is_followed?community_id=${communityId}`);
  }

  async getFollowedCommunities() {
    return this.request<CommunitySummary[]>("/followed_communities");
  }

  async getCommunityDetail(id: string) {
    return this.request<CommunityDetail>(`/community/${id}`);
  }

  async getPosts(params?: {
    page?: number;
    size?: number;
    community_id?: string;
    user_name?: string;
    sort_by?: string;
    order?: string;
  }) {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.size) searchParams.set("size", params.size.toString());
    if (params?.community_id) searchParams.set("community_id", params.community_id);
    if (params?.user_name) searchParams.set("user_name", params.user_name);
    if (params?.sort_by) searchParams.set("sort_by", params.sort_by);
    if (params?.order) searchParams.set("order", params.order);

    const query = searchParams.toString();
    return this.request<PostsResponse | PostListItem[]>(`/posts${query ? `?${query}` : ""}`);
  }

  async getUserPosts(
    id: string,
    params?: {
      page?: number;
      size?: number;
      sort_by?: string;
      order?: string;
    }
  ) {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.size) searchParams.set("size", params.size.toString());
    if (params?.sort_by) searchParams.set("sort_by", params.sort_by);
    if (params?.order) searchParams.set("order", params.order);

    const query = searchParams.toString();
    return this.request<PostsResponse | PostListItem[]>(`/users/${id}/posts${query ? `?${query}` : ""}`);
  }

  async getPostDetail(id: string) {
    return this.request<PostDetail>(`/post_detail/${id}`);
  }

  async createPost(title: string, content: string, communityId: string) {
    return this.request("/create_post", {
      method: "POST",
      body: JSON.stringify({
        title,
        content,
        community_id: Number(communityId),
      }),
    });
  }

  async vote(postId: string, direction: number) {
    return this.request("/vote", {
      method: "POST",
      body: JSON.stringify({
        post_id: postId,
        direction: direction.toString(),
      }),
    });
  }
}

export const apiClient = new ApiClient(API_BASE_URL);
export type { ApiResponse };
