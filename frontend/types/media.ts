export interface MediaFile {
  id: string;
  bucket: string;
  object_key: string;
  file_name: string;
  mime_type: string;
  size: number;
  checksum: string;
  source_service?: string;
  source_module?: string;
  uploaded_by: string;
  is_public: boolean;
  ref_count: number;
  url?: string;
  created_at: string;
  deleted_at?: string;
}

export interface MediaReference {
  id: number;
  media_id: string;
  ref_service: string;
  ref_table: string;
  ref_id: string;
  ref_field: string;
  created_at: string;
}

export interface UsageResult {
  media_id: string;
  ref_count: number;
  references: MediaReference[];
}

export interface ListResult {
  items: MediaFile[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
  has_more?: boolean;
  next_cursor?: string;
}

export interface FilterOptions {
  source_services: string[];
  source_modules: string[];
}

export interface ListParams {
  pagination?: "offset" | "cursor";
  cursor?: string;
  page?: number;
  limit?: number;
  type?: string;
  search?: string;
  uploaded_by?: string;
  source_service?: string;
  source_module?: string;
  sort_by?: string;
  sort_dir?: string;
}

export interface AddReferenceInput {
  media_id: string;
  ref_service: string;
  ref_table: string;
  ref_id: string;
  ref_field: string;
}

export interface RemoveReferenceInput {
  media_id: string;
  ref_service: string;
  ref_table: string;
  ref_id: string;
  ref_field: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
}

export interface CreateUserRequest {
  username: string;
  password: string;
  role?: "admin" | "editor" | "viewer";
}

export interface UpdateUserRequest {
  role?: "admin" | "editor" | "viewer";
  password?: string;
}

export interface UserSummary {
  id: string;
  username: string;
  role: string;
  created_at: string;
}

export interface CreateProjectRequest {
  name: string;
  scopes?: string[];
  upload_policy?: ProjectUploadPolicy;
}

export interface UpdateProjectRequest {
  name?: string;
  scopes?: string[];
  upload_policy?: ProjectUploadPolicy;
  is_active?: boolean;
}

export interface ProjectUploadPolicy {
  limits_mb: Record<"image" | "video" | "audio" | "document" | "archive" | "other", number>;
}

export interface CreateProjectResponse {
  id: string;
  name: string;
  client_id: string;
  client_secret: string;
  scopes: string[];
  upload_policy: ProjectUploadPolicy;
}

export interface ProjectSummary {
  id: string;
  name: string;
  client_id: string;
  scopes: string[];
  upload_policy: ProjectUploadPolicy;
  is_active: boolean;
  created_at: string;
}

export interface ProjectUploadLog {
  id: number;
  project_id: string;
  media_id?: string;
  file_name: string;
  mime_type: string;
  size: number;
  source_service?: string;
  source_module?: string;
  status: "success" | "failed";
  error_message?: string;
  uploaded_by: string;
  created_at: string;
}

export type MimeGroup = "image" | "video" | "audio" | "document" | "other" | "";

export interface UploadProgress {
  file: File;
  progress: number;
  status: "pending" | "uploading" | "done" | "error";
  result?: MediaFile;
  error?: string;
}
