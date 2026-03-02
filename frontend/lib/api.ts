import axios from "axios";
import type {
  AddReferenceInput,
  CreateProjectRequest,
  CreateProjectResponse,
  CreateUserRequest,
  FilterOptions,
  ListParams,
  ListResult,
  LoginRequest,
  LoginResponse,
  MediaFile,
  ProjectSummary,
  ProjectUploadLog,
  RemoveReferenceInput,
  UpdateProjectRequest,
  UpdateUserRequest,
  UserSummary,
  UsageResult,
} from "@/types/media";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export const apiClient = axios.create({
  baseURL: BASE_URL,
  timeout: 60_000,
});

// Attach JWT token to every request
apiClient.interceptors.request.use((config) => {
  const token =
    typeof window !== "undefined" ? localStorage.getItem("auth_token") : null;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// ─── Media API ────────────────────────────────────────────────────────────────

export const mediaApi = {
  /**
   * Upload a file with streaming progress tracking.
   * onProgress: 0–100
   */
  upload(
    file: File,
    opts: {
      sourceService?: string;
      sourceModule?: string;
      isPublic?: boolean;
      onProgress?: (pct: number) => void;
    } = {}
  ): Promise<MediaFile> {
    const form = new FormData();
    if (opts.sourceService) form.append("source_service", opts.sourceService);
    if (opts.sourceModule) form.append("source_module", opts.sourceModule);
    form.append("is_public", opts.isPublic ? "true" : "false");
    form.append("file", file); // file must be last – handler streams it immediately

    return apiClient
      .post<MediaFile>("/media/upload", form, {
        headers: { "Content-Type": "multipart/form-data" },
        onUploadProgress: (e) => {
          if (e.total) {
            opts.onProgress?.(Math.round((e.loaded / e.total) * 100));
          }
        },
      })
      .then((r) => r.data);
  },

  list(params: ListParams = {}): Promise<ListResult> {
    return apiClient
      .get<ListResult>("/media", { params })
      .then((r) => r.data);
  },

  getFilterOptions(): Promise<FilterOptions> {
    return apiClient.get<FilterOptions>("/media/filter-options").then((r) => r.data);
  },

  getById(id: string): Promise<MediaFile> {
    return apiClient.get<MediaFile>(`/media/${id}`).then((r) => r.data);
  },

  delete(id: string): Promise<void> {
    return apiClient.delete(`/media/${id}`).then(() => undefined);
  },

  getUsage(id: string): Promise<UsageResult> {
    return apiClient.get<UsageResult>(`/media/${id}/usage`).then((r) => r.data);
  },
};

// ─── Reference API ────────────────────────────────────────────────────────────

export const referenceApi = {
  add(input: AddReferenceInput): Promise<void> {
    return apiClient.post("/media/reference", input).then(() => undefined);
  },

  remove(input: RemoveReferenceInput): Promise<void> {
    return apiClient.delete("/media/reference", { data: input }).then(() => undefined);
  },
};

// ─── Auth API ────────────────────────────────────────────────────────────────

export const authApi = {
  login(input: LoginRequest): Promise<LoginResponse> {
    return apiClient.post<LoginResponse>("/auth/login", input).then((r) => r.data);
  },

  listUsers(): Promise<UserSummary[]> {
    return apiClient.get<{ items: UserSummary[] }>("/auth/users").then((r) => r.data.items);
  },

  createUser(input: CreateUserRequest): Promise<UserSummary> {
    return apiClient.post<UserSummary>("/auth/users", input).then((r) => r.data);
  },

  updateUser(id: string, input: UpdateUserRequest): Promise<UserSummary> {
    return apiClient.put<UserSummary>(`/auth/users/${id}`, input).then((r) => r.data);
  },

  deleteUser(id: string): Promise<void> {
    return apiClient.delete(`/auth/users/${id}`).then(() => undefined);
  },

  listProjects(): Promise<ProjectSummary[]> {
    return apiClient.get<{ items: ProjectSummary[] }>("/auth/projects").then((r) => r.data.items);
  },

  createProject(input: CreateProjectRequest): Promise<CreateProjectResponse> {
    return apiClient.post<CreateProjectResponse>("/auth/projects", input).then((r) => r.data);
  },

  updateProject(id: string, input: UpdateProjectRequest): Promise<ProjectSummary> {
    return apiClient.put<ProjectSummary>(`/auth/projects/${id}`, input).then((r) => r.data);
  },

  deleteProject(id: string): Promise<void> {
    return apiClient.delete(`/auth/projects/${id}`).then(() => undefined);
  },

  listProjectUploadLogs(projectId: string, limit = 20): Promise<ProjectUploadLog[]> {
    return apiClient
      .get<{ items: ProjectUploadLog[] }>(`/auth/projects/${projectId}/upload-logs`, {
        params: { limit },
      })
      .then((r) => r.data.items);
  },

  issueProjectToken(clientId: string, clientSecret: string): Promise<LoginResponse> {
    return apiClient
      .post<LoginResponse>("/auth/project-token", {
        client_id: clientId,
        client_secret: clientSecret,
      })
      .then((r) => r.data);
  },
};
