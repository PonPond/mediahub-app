"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { authApi } from "@/lib/api";
import { getErrorMessage } from "@/lib/error";
import { toast } from "@/components/ui/use-toast";
import type { CreateProjectResponse, ProjectSummary, ProjectUploadLog, ProjectUploadPolicy, UserSummary } from "@/types/media";
import { formatBytes, formatDate } from "@/lib/utils";
import { AlertTriangle, Check, Copy, FolderKey, HelpCircle, KeyRound, ListFilter, MoreVertical, PlusCircle, Shield, Trash2, Users } from "lucide-react";

const AVAILABLE_SCOPES = ["media:read", "media:write", "reference:write"] as const;
const POLICY_GROUPS = [
  {
    key: "image",
    label: "Images",
    tooltip: "MIME ที่ขึ้นต้นด้วย image/ เช่น image/jpeg, image/png",
  },
  {
    key: "video",
    label: "Videos",
    tooltip: "MIME ที่ขึ้นต้นด้วย video/ เช่น video/mp4, video/webm",
  },
  {
    key: "audio",
    label: "Audio",
    tooltip: "MIME ที่ขึ้นต้นด้วย audio/ เช่น audio/mpeg, audio/wav",
  },
  {
    key: "document",
    label: "Documents",
    tooltip:
      "PDF / text/* / MIME ที่มี document|spreadsheet|presentation และนามสกุลเช่น .sql .csv .docx .xlsx .pptx",
  },
  {
    key: "archive",
    label: "Archives",
    tooltip: "MIME หรือไฟล์นามสกุล .zip .tar .gz .tgz .7z .rar",
  },
  {
    key: "other",
    label: "Other",
    tooltip: "ไฟล์ที่ไม่เข้าเงื่อนไขของ image/video/audio/document/archive",
  },
] as const;
const SIZE_PRESETS_MB = [0, 1, 2, 5, 10, 20, 30, 50, 100, 200, 500, 1024];

type UploadGroupKey = (typeof POLICY_GROUPS)[number]["key"];

export function AuthAdminPanel() {
  const [activeMenu, setActiveMenu] = useState<"users" | "projects">("users");
  const [userView, setUserView] = useState<"create" | "list">("create");
  const [projectView, setProjectView] = useState<"create" | "list">("create");
  const [users, setUsers] = useState<UserSummary[]>([]);
  const [projects, setProjects] = useState<ProjectSummary[]>([]);
  const [userFilter, setUserFilter] = useState("");
  const [projectFilter, setProjectFilter] = useState("");

  const [newUsername, setNewUsername] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [newRole, setNewRole] = useState<"admin" | "editor" | "viewer">("editor");

  const [projectName, setProjectName] = useState("");
  const [projectScopes, setProjectScopes] = useState<string[]>([...AVAILABLE_SCOPES]);
  const [projectUploadPolicy, setProjectUploadPolicy] = useState<ProjectUploadPolicy>(createDefaultUploadPolicy());
  const [createdProject, setCreatedProject] = useState<CreateProjectResponse | null>(null);

  const [tokenClientID, setTokenClientID] = useState("");
  const [tokenClientSecret, setTokenClientSecret] = useState("");
  const [issuedToken, setIssuedToken] = useState("");
  const [copiedField, setCopiedField] = useState<"client_id" | "client_secret" | null>(null);

  const [editingUserId, setEditingUserId] = useState<string | null>(null);
  const [editUserRole, setEditUserRole] = useState<"admin" | "editor" | "viewer">("editor");
  const [editUserPassword, setEditUserPassword] = useState("");

  const [editingProjectId, setEditingProjectId] = useState<string | null>(null);
  const [editProjectName, setEditProjectName] = useState("");
  const [editProjectScopes, setEditProjectScopes] = useState<string[]>([]);
  const [editProjectUploadPolicy, setEditProjectUploadPolicy] = useState<ProjectUploadPolicy>(createDefaultUploadPolicy());
  const [editProjectActive, setEditProjectActive] = useState(true);
  const [deleteTarget, setDeleteTarget] = useState<{ type: "user" | "project"; id: string; label: string } | null>(null);
  const [isDeletingUser, setIsDeletingUser] = useState(false);
  const [isDeletingProject, setIsDeletingProject] = useState(false);
  const [expandedProjectLogsId, setExpandedProjectLogsId] = useState<string | null>(null);
  const [projectLogsByProject, setProjectLogsByProject] = useState<Record<string, ProjectUploadLog[]>>({});
  const [projectLogsLoadingId, setProjectLogsLoadingId] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const [u, p] = await Promise.all([authApi.listUsers(), authApi.listProjects()]);
      setUsers(u);
      setProjects(p);
    } catch {
      // Non-admin token may not have access.
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const createUser = async () => {
    try {
      const u = await authApi.createUser({
        username: newUsername,
        password: newPassword,
        role: newRole,
      });
      setUsers((prev) => [u, ...prev]);
      setNewUsername("");
      setNewPassword("");
      toast({ title: "User created", description: u.username });
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to create user.");
      toast({ title: "Create user failed", description: msg, variant: "destructive" });
    }
  };

  const createProject = async () => {
    try {
      const p = await authApi.createProject({
        name: projectName,
        scopes: projectScopes,
        upload_policy: projectUploadPolicy,
      });
      setCreatedProject(p);
      setTokenClientID(p.client_id);
      setTokenClientSecret(p.client_secret);
      setProjectName("");
      setProjectUploadPolicy(createDefaultUploadPolicy());
      await load();
      toast({ title: "Project credential created", description: p.name });
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to create project.");
      toast({ title: "Create project failed", description: msg, variant: "destructive" });
    }
  };

  const exchangeToken = async () => {
    try {
      const clientId = tokenClientID || createdProject?.client_id || "";
      const clientSecret = tokenClientSecret || createdProject?.client_secret || "";
      const t = await authApi.issueProjectToken(clientId, clientSecret);
      setIssuedToken(t.access_token);
      toast({ title: "Project token issued" });
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to issue project token.");
      toast({ title: "Issue token failed", description: msg, variant: "destructive" });
    }
  };

  const startEditUser = (u: UserSummary) => {
    setEditingUserId(u.id);
    setEditUserRole((u.role as "admin" | "editor" | "viewer") || "editor");
    setEditUserPassword("");
  };

  const saveUserEdit = async () => {
    if (!editingUserId) return;
    try {
      const updated = await authApi.updateUser(editingUserId, {
        role: editUserRole,
        password: editUserPassword || undefined,
      });
      setUsers((prev) => prev.map((u) => (u.id === updated.id ? updated : u)));
      setEditingUserId(null);
      setEditUserPassword("");
      toast({ title: "User updated" });
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to update user.");
      toast({ title: "Update user failed", description: msg, variant: "destructive" });
    }
  };

  const removeUser = async (id: string) => {
    setIsDeletingUser(true);
    try {
      await authApi.deleteUser(id);
      setUsers((prev) => prev.filter((u) => u.id !== id));
      setDeleteTarget(null);
      toast({ title: "User deleted" });
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to delete user.");
      toast({ title: "Delete user failed", description: msg, variant: "destructive" });
    } finally {
      setIsDeletingUser(false);
    }
  };

  const startEditProject = (p: ProjectSummary) => {
    setEditingProjectId(p.id);
    setEditProjectName(p.name);
    setEditProjectScopes(p.scopes);
    setEditProjectUploadPolicy(normalizeUploadPolicy(p.upload_policy));
    setEditProjectActive(!!p.is_active);
  };

  const saveProjectEdit = async () => {
    if (!editingProjectId) return;
    try {
      const updated = await authApi.updateProject(editingProjectId, {
        name: editProjectName,
        scopes: editProjectScopes,
        upload_policy: editProjectUploadPolicy,
        is_active: editProjectActive,
      });
      setProjects((prev) => prev.map((p) => (p.id === updated.id ? updated : p)));
      setEditingProjectId(null);
      toast({ title: "Project updated" });
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to update project.");
      toast({ title: "Update project failed", description: msg, variant: "destructive" });
    }
  };

  const removeProject = async (id: string) => {
    setIsDeletingProject(true);
    try {
      await authApi.deleteProject(id);
      setProjects((prev) => prev.filter((p) => p.id !== id));
      setDeleteTarget(null);
      toast({ title: "Project deleted" });
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to delete project.");
      toast({ title: "Delete project failed", description: msg, variant: "destructive" });
    } finally {
      setIsDeletingProject(false);
    }
  };

  const toggleProjectLogs = async (projectId: string) => {
    if (expandedProjectLogsId === projectId) {
      setExpandedProjectLogsId(null);
      return;
    }

    setExpandedProjectLogsId(projectId);
    if (projectLogsByProject[projectId]) return;

    setProjectLogsLoadingId(projectId);
    try {
      const logs = await authApi.listProjectUploadLogs(projectId, 20);
      setProjectLogsByProject((prev) => ({ ...prev, [projectId]: logs }));
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to load upload activity.");
      toast({ title: "Load upload activity failed", description: msg, variant: "destructive" });
    } finally {
      setProjectLogsLoadingId((prev) => (prev === projectId ? null : prev));
    }
  };

  const copyCredential = async (field: "client_id" | "client_secret", value: string) => {
    if (!value) return;
    try {
      await navigator.clipboard.writeText(value);
    } catch {
      const textarea = document.createElement("textarea");
      textarea.value = value;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand("copy");
      document.body.removeChild(textarea);
    }
    setCopiedField(field);
    setTimeout(() => setCopiedField((prev) => (prev === field ? null : prev)), 1800);
    toast({ title: "Copied", description: field === "client_id" ? "client_id copied" : "client_secret copied" });
  };

  const filteredUsers = users.filter((u) =>
    `${u.username} ${u.role}`.toLowerCase().includes(userFilter.toLowerCase())
  );

  const filteredProjects = projects.filter((p) =>
    `${p.name} ${p.client_id} ${p.scopes.join(" ")} ${p.is_active ? "active" : "inactive"}`
      .toLowerCase()
      .includes(projectFilter.toLowerCase())
  );

  const menuButtonClass = (isActive: boolean) =>
    `w-full flex items-center gap-2 px-3 py-2.5 rounded-xl text-sm text-left transition ${
      isActive
        ? "bg-secondary text-foreground border border-border"
        : "text-muted-foreground hover:bg-muted hover:text-foreground border border-transparent"
    }`;

  const submenuButtonClass = (isActive: boolean) =>
    `w-full text-left rounded-lg px-3 py-2 text-sm transition ${
      isActive ? "bg-secondary/80 text-foreground" : "text-muted-foreground hover:bg-muted hover:text-foreground"
    }`;

  const rolePillClass = (role: string) => {
    if (role === "admin") return "border-red-500/30 bg-red-500/10 text-red-400 dark:text-red-300";
    if (role === "editor") return "border-sky-500/30 bg-sky-500/10 text-sky-600 dark:text-sky-300";
    return "border-border bg-muted text-muted-foreground";
  };

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-[280px,1fr]">
      <aside className="h-fit rounded-3xl border border-border bg-card p-4 lg:sticky lg:top-20">
        <p className="px-2 pb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Auth Admin</p>
        <nav className="space-y-2">
          <button
            className={menuButtonClass(activeMenu === "users")}
            onClick={() => {
              setActiveMenu("users");
              setUserView("create");
            }}
          >
            <Users className="h-4 w-4" />
            CMS Users
          </button>
          {activeMenu === "users" && (
            <div className="ml-3 border-l border-border pl-3 space-y-1">
              <button className={submenuButtonClass(userView === "create")} onClick={() => setUserView("create")}>
                Create User
              </button>
              <button className={submenuButtonClass(userView === "list")} onClick={() => setUserView("list")}>
                User List
              </button>
            </div>
          )}

          <button
            className={menuButtonClass(activeMenu === "projects")}
            onClick={() => {
              setActiveMenu("projects");
              setProjectView("create");
            }}
          >
            <FolderKey className="h-4 w-4" />
            Project Credentials
          </button>
          {activeMenu === "projects" && (
            <div className="ml-3 border-l border-border pl-3 space-y-1">
              <button
                className={submenuButtonClass(projectView === "create")}
                onClick={() => setProjectView("create")}
              >
                Create Credential
              </button>
              <button
                className={submenuButtonClass(projectView === "list")}
                onClick={() => setProjectView("list")}
              >
                Project List
              </button>
            </div>
          )}
        </nav>
      </aside>

      <div className="space-y-4">
        {activeMenu === "users" && userView === "create" && (
          <section className="rounded-2xl border border-border bg-card/70 p-5 space-y-5">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <Shield className="h-5 w-5 text-primary" />
                <h3 className="font-semibold text-foreground">Create CMS User</h3>
              </div>
              <span className="rounded-full border border-border bg-muted px-3 py-1 text-xs text-muted-foreground">
                Auth Management
              </span>
            </div>
            <div className="grid gap-3 rounded-xl border border-border bg-card p-4 sm:grid-cols-4">
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Username</p>
                <Input
                  value={newUsername}
                  onChange={(e) => setNewUsername(e.target.value)}
                  placeholder="username"
                />
              </div>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Password</p>
                <Input
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="min 8 chars"
                />
              </div>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Role</p>
                <select
                  className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm text-foreground"
                  value={newRole}
                  onChange={(e) => setNewRole(e.target.value as "admin" | "editor" | "viewer")}
                >
                  <option value="admin">admin</option>
                  <option value="editor">editor</option>
                  <option value="viewer">viewer</option>
                </select>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">Action</p>
                <Button
                  onClick={createUser}
                  disabled={!newUsername || newPassword.length < 8}
                  className="w-full"
                >
                  <PlusCircle className="mr-2 h-4 w-4" />
                  Create User
                </Button>
              </div>
            </div>
            <p className="text-xs text-muted-foreground">
              Tip: use User List from the left menu to edit role, reset password, and delete user.
            </p>
          </section>
        )}

        {activeMenu === "users" && userView === "list" && (
          <section className="rounded-2xl border border-border bg-card/70 p-5">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <h3 className="font-semibold text-foreground">CMS User List</h3>
              <div className="flex w-full max-w-sm items-center gap-2">
                <ListFilter className="h-4 w-4 text-muted-foreground" />
                <Input
                  value={userFilter}
                  onChange={(e) => setUserFilter(e.target.value)}
                  placeholder="Filter username, role..."
                />
              </div>
            </div>
            <div className="overflow-x-auto rounded-xl border border-border bg-muted/20">
              <div className="min-w-[980px]">
                <div className="grid grid-cols-12 gap-2 border-b border-border bg-secondary/50 px-4 py-3 text-sm font-medium text-foreground">
                  <div className="col-span-4">Username</div>
                  <div className="col-span-2">Role</div>
                  <div className="col-span-2">Status</div>
                  <div className="col-span-3">Created</div>
                  <div className="col-span-1 text-right">Action</div>
                </div>
                {filteredUsers.map((u) => (
                  <div key={u.id} className="grid grid-cols-12 gap-2 border-b border-border/50 px-4 py-3 text-sm text-foreground">
                    <div className="col-span-4">
                      <div className="font-medium">{u.username}</div>
                      {editingUserId === u.id && (
                        <Input
                          className="mt-2 h-8"
                          value={editUserPassword}
                          onChange={(e) => setEditUserPassword(e.target.value)}
                          placeholder="new password (optional)"
                          type="password"
                        />
                      )}
                    </div>
                    <div className="col-span-2">
                      {editingUserId === u.id ? (
                        <select
                          className="h-8 w-full rounded-md border border-input bg-background px-2 text-xs text-foreground"
                          value={editUserRole}
                          onChange={(e) =>
                            setEditUserRole(e.target.value as "admin" | "editor" | "viewer")
                          }
                        >
                          <option value="admin">admin</option>
                          <option value="editor">editor</option>
                          <option value="viewer">viewer</option>
                        </select>
                      ) : (
                        <span className={`inline-flex rounded-full border px-2.5 py-1 text-xs ${rolePillClass(u.role)}`}>
                          {u.role}
                        </span>
                      )}
                    </div>
                    <div className="col-span-2">
                      <span className="inline-flex rounded-full border border-emerald-500/30 bg-emerald-500/10 px-2.5 py-1 text-xs text-emerald-600 dark:text-emerald-300">
                        active
                      </span>
                    </div>
                    <div className="col-span-3 text-muted-foreground">
                      {formatDate(u.created_at)}
                    </div>
                    <div className="col-span-1 flex justify-end gap-1">
                      {editingUserId === u.id ? (
                        <Button size="sm" className="h-8 px-3" onClick={saveUserEdit}>
                          Save
                        </Button>
                      ) : (
                        <details className="relative">
                          <summary className="flex h-8 w-8 cursor-pointer list-none items-center justify-center rounded-md border border-border bg-muted text-muted-foreground hover:bg-secondary">
                            <MoreVertical className="h-4 w-4" />
                          </summary>
                          <div className="absolute right-0 top-9 z-20 w-36 rounded-lg border border-border bg-popover p-1.5 shadow-xl">
                            <button
                              className="w-full rounded-md px-2 py-1.5 text-left text-sm text-popover-foreground hover:bg-accent"
                              onClick={() => startEditUser(u)}
                            >
                              Edit
                            </button>
                            <button
                              className="w-full rounded-md px-2 py-1.5 text-left text-sm text-red-500 hover:bg-accent"
                              onClick={() => setDeleteTarget({ type: "user", id: u.id, label: u.username })}
                            >
                              Delete
                            </button>
                          </div>
                        </details>
                      )}
                    </div>
                  </div>
                ))}
                {filteredUsers.length === 0 && <p className="px-4 py-5 text-sm text-muted-foreground">No users</p>}
              </div>
            </div>
          </section>
        )}

        {activeMenu === "projects" && projectView === "create" && (
          <section className="rounded-2xl border bg-card p-5 space-y-4">
            <div className="flex items-center gap-2">
              <KeyRound className="h-5 w-5 text-primary" />
              <h3 className="font-semibold">Create Project Credential</h3>
            </div>
            <p className="text-xs text-muted-foreground">
              Client ID และ Client Secret จะถูกสุ่มให้อัตโนมัติเมื่อกด Create Credential
            </p>
            <div className="grid gap-2 sm:grid-cols-3">
              <Input
                value={projectName}
                onChange={(e) => setProjectName(e.target.value)}
                placeholder="Project name"
              />
              <ScopeMultiSelect value={projectScopes} onChange={setProjectScopes} />
              <Button onClick={createProject} disabled={!projectName || projectScopes.length === 0}>
                <PlusCircle className="mr-2 h-4 w-4" />
                Create Credential
              </Button>
            </div>
            <div className="rounded-md border bg-muted/30 p-3">
              <p className="mb-2 text-xs font-medium text-muted-foreground">Per-file-type upload limits</p>
              <UploadPolicyEditor value={projectUploadPolicy} onChange={setProjectUploadPolicy} />
            </div>

            {createdProject && (
              <div className="space-y-1 rounded-md border bg-muted/40 p-3 text-xs">
                <div className="flex flex-wrap items-center gap-2">
                  <strong>client_id:</strong>
                  <span className="font-mono break-all">{createdProject.client_id}</span>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="h-6 px-2 text-[11px]"
                    onClick={() => copyCredential("client_id", createdProject.client_id)}
                  >
                    {copiedField === "client_id" ? (
                      <>
                        <Check className="mr-1 h-3 w-3" />
                        Copied
                      </>
                    ) : (
                      <>
                        <Copy className="mr-1 h-3 w-3" />
                        Copy
                      </>
                    )}
                  </Button>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <strong>client_secret:</strong>
                  <span className="font-mono break-all">{createdProject.client_secret}</span>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="h-6 px-2 text-[11px]"
                    onClick={() => copyCredential("client_secret", createdProject.client_secret)}
                  >
                    {copiedField === "client_secret" ? (
                      <>
                        <Check className="mr-1 h-3 w-3" />
                        Copied
                      </>
                    ) : (
                      <>
                        <Copy className="mr-1 h-3 w-3" />
                        Copy
                      </>
                    )}
                  </Button>
                </div>
                <p className="text-destructive">Save client_secret now. It is shown only once.</p>
              </div>
            )}

            <h4 className="pt-1 text-sm font-medium">Exchange Project Token</h4>
            <div className="grid gap-2 sm:grid-cols-3">
              <Input
                value={tokenClientID || createdProject?.client_id || ""}
                placeholder="client_id (auto)"
                readOnly
              />
              <Input
                value={tokenClientSecret || createdProject?.client_secret || ""}
                placeholder="client_secret (auto)"
                readOnly
              />
              <Button
                onClick={exchangeToken}
                disabled={!(tokenClientID || createdProject?.client_id) || !(tokenClientSecret || createdProject?.client_secret)}
              >
                Exchange Token
              </Button>
            </div>

            {issuedToken && (
              <textarea
                className="w-full rounded-md border bg-background p-2 text-xs font-mono text-foreground"
                rows={4}
                value={issuedToken}
                readOnly
              />
            )}
          </section>
        )}

        {activeMenu === "projects" && projectView === "list" && (
          <section className="rounded-2xl border border-border bg-card/70 p-5">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <h3 className="font-semibold text-foreground">Project Credential List</h3>
              <div className="flex w-full max-w-sm items-center gap-2">
                <ListFilter className="h-4 w-4 text-muted-foreground" />
                <Input
                  value={projectFilter}
                  onChange={(e) => setProjectFilter(e.target.value)}
                  placeholder="Filter name, client id, scope..."
                />
              </div>
            </div>
            <div className="overflow-x-auto rounded-xl border border-border bg-muted/20">
              <div className="min-w-[1100px]">
                <div className="grid grid-cols-12 gap-2 border-b border-border bg-secondary/50 px-4 py-3 text-sm font-medium text-foreground">
                  <div className="col-span-2">Name</div>
                  <div className="col-span-3">Client ID</div>
                  <div className="col-span-3">Scopes</div>
                  <div className="col-span-1">Status</div>
                  <div className="col-span-2">Created</div>
                  <div className="col-span-1 text-right">Action</div>
                </div>
                {filteredProjects.map((p) => (
                  <div key={p.id} className="border-b border-border/50">
                    <div className="grid grid-cols-12 gap-2 px-4 py-3 text-sm text-foreground">
                      <div className="col-span-2">
                        {editingProjectId === p.id ? (
                          <Input
                            className="h-8"
                            value={editProjectName}
                            onChange={(e) => setEditProjectName(e.target.value)}
                          />
                        ) : (
                          <div className="space-y-1">
                            <div className="font-medium">{p.name}</div>
                            <button
                              className="text-[11px] text-sky-500 hover:text-sky-400"
                              onClick={() => void toggleProjectLogs(p.id)}
                            >
                              {expandedProjectLogsId === p.id ? "Hide activity" : "Recent upload activity"}
                            </button>
                          </div>
                        )}
                      </div>
                      <div className="col-span-3 truncate font-mono text-muted-foreground">{p.client_id}</div>
                      <div className="col-span-3">
                        {editingProjectId === p.id ? (
                          <div className="space-y-2">
                            <ScopeMultiSelect value={editProjectScopes} onChange={setEditProjectScopes} />
                            <UploadPolicyEditor
                              value={editProjectUploadPolicy}
                              onChange={setEditProjectUploadPolicy}
                              compact
                            />
                          </div>
                        ) : (
                          <div className="space-y-1">
                            <span className="text-foreground">{p.scopes.join(", ")}</span>
                            <p className="text-[11px] text-muted-foreground">{uploadPolicySummary(p.upload_policy)}</p>
                          </div>
                        )}
                      </div>
                      <div className="col-span-1">
                        {editingProjectId === p.id ? (
                          <label className="inline-flex items-center gap-1 text-[11px] text-foreground">
                            <input
                              type="checkbox"
                              checked={editProjectActive}
                              onChange={(e) => setEditProjectActive(e.target.checked)}
                            />
                            active
                          </label>
                        ) : (
                          <span
                            className={`inline-flex rounded-full border px-2.5 py-1 text-xs ${
                              p.is_active
                                ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-300"
                                : "border-border bg-muted text-muted-foreground"
                            }`}
                          >
                            {p.is_active ? "active" : "inactive"}
                          </span>
                        )}
                      </div>
                      <div className="col-span-2 text-muted-foreground">{formatDate(p.created_at)}</div>
                      <div className="col-span-1 flex justify-end gap-1">
                        {editingProjectId === p.id ? (
                          <Button size="sm" className="h-8 px-3" onClick={saveProjectEdit}>
                            Save
                          </Button>
                        ) : (
                          <details className="relative">
                            <summary className="flex h-8 w-8 cursor-pointer list-none items-center justify-center rounded-md border border-border bg-muted text-muted-foreground hover:bg-secondary">
                              <MoreVertical className="h-4 w-4" />
                            </summary>
                            <div className="absolute right-0 top-9 z-20 w-36 rounded-lg border border-border bg-popover p-1.5 shadow-xl">
                              <button
                                className="w-full rounded-md px-2 py-1.5 text-left text-sm text-popover-foreground hover:bg-accent"
                                onClick={() => startEditProject(p)}
                              >
                                Edit
                              </button>
                              <button
                                className="w-full rounded-md px-2 py-1.5 text-left text-sm text-red-500 hover:bg-accent"
                                onClick={() => setDeleteTarget({ type: "project", id: p.id, label: p.name })}
                              >
                                Delete
                              </button>
                            </div>
                          </details>
                        )}
                      </div>
                    </div>
                    {expandedProjectLogsId === p.id && (
                      <div className="px-4 pb-4">
                        <div className="rounded-lg border border-border bg-muted/40 p-3">
                          <p className="mb-2 text-xs font-medium text-foreground">Recent Upload Activity (latest 20)</p>
                          {projectLogsLoadingId === p.id ? (
                            <p className="text-xs text-muted-foreground">Loading...</p>
                          ) : (projectLogsByProject[p.id] ?? []).length === 0 ? (
                            <p className="text-xs text-muted-foreground">No upload logs yet.</p>
                          ) : (
                            <div className="space-y-2">
                              {(projectLogsByProject[p.id] ?? []).map((log) => (
                                <div key={log.id} className="grid grid-cols-12 gap-2 rounded-md border border-border bg-muted/30 px-2 py-2 text-[11px] text-foreground">
                                  <div className="col-span-3 truncate" title={log.file_name}>
                                    {log.file_name}
                                  </div>
                                  <div className="col-span-2 truncate text-muted-foreground">{log.mime_type}</div>
                                  <div className="col-span-1 text-muted-foreground">{formatBytes(log.size || 0, 1)}</div>
                                  <div className="col-span-2 truncate text-muted-foreground">
                                    {(log.source_service || "-") + "/" + (log.source_module || "-")}
                                  </div>
                                  <div className="col-span-1">
                                    <span
                                      className={`inline-flex rounded-full border px-2 py-0.5 ${
                                        log.status === "success"
                                          ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-300"
                                          : "border-red-500/30 bg-red-500/10 text-red-500 dark:text-red-300"
                                      }`}
                                    >
                                      {log.status}
                                    </span>
                                  </div>
                                  <div className="col-span-3 truncate text-muted-foreground" title={log.error_message || ""}>
                                    {log.error_message || formatDate(log.created_at)}
                                  </div>
                                </div>
                              ))}
                            </div>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                ))}
                {filteredProjects.length === 0 && <p className="px-4 py-5 text-sm text-muted-foreground">No projects</p>}
              </div>
            </div>
          </section>
        )}
      </div>

      <AuthDeleteConfirmDialog
        target={deleteTarget}
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        onConfirm={() => {
          if (!deleteTarget) return;
          if (deleteTarget.type === "user") {
            void removeUser(deleteTarget.id);
            return;
          }
          void removeProject(deleteTarget.id);
        }}
        isDeleting={isDeletingUser || isDeletingProject}
      />
    </div>
  );
}

function createDefaultUploadPolicy(): ProjectUploadPolicy {
  return {
    limits_mb: {
      image: 10,
      video: 200,
      audio: 30,
      document: 20,
      archive: 50,
      other: 5,
    },
  };
}

function normalizeUploadPolicy(policy?: ProjectUploadPolicy): ProjectUploadPolicy {
  const base = createDefaultUploadPolicy();
  if (!policy?.limits_mb) return base;
  const next = { ...base.limits_mb };
  for (const group of POLICY_GROUPS) {
    const v = policy.limits_mb[group.key];
    if (Number.isFinite(v) && v >= 0) {
      next[group.key] = Math.floor(v);
    }
  }
  return { limits_mb: next };
}

function uploadPolicySummary(policy?: ProjectUploadPolicy): string {
  const p = normalizeUploadPolicy(policy);
  return POLICY_GROUPS.map((g) => `${g.label}: ${p.limits_mb[g.key]}MB`).join(" | ");
}

function UploadPolicyEditor({
  value,
  onChange,
  compact = false,
}: {
  value: ProjectUploadPolicy;
  onChange: (next: ProjectUploadPolicy) => void;
  compact?: boolean;
}) {
  const policy = normalizeUploadPolicy(value);

  const setLimit = (group: UploadGroupKey, mb: number) => {
    onChange({
      limits_mb: {
        ...policy.limits_mb,
        [group]: mb,
      },
    });
  };

  return (
    <div className={compact ? "grid gap-2 sm:grid-cols-2" : "grid gap-2 sm:grid-cols-3"}>
      {POLICY_GROUPS.map((group) => (
        <label key={group.key} className="space-y-1">
          <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
            {group.label}
            <InlineTooltip text={group.tooltip} />
          </span>
          <select
            className="h-9 w-full rounded-md border border-input bg-background px-2 text-xs text-foreground"
            value={String(policy.limits_mb[group.key])}
            onChange={(e) => setLimit(group.key, Number(e.target.value))}
          >
            {SIZE_PRESETS_MB.map((mb) => (
              <option key={`${group.key}-${mb}`} value={mb}>
                {mb === 0 ? "Disabled" : `${mb} MB`}
              </option>
            ))}
          </select>
        </label>
      ))}
    </div>
  );
}

function InlineTooltip({ text }: { text: string }) {
  return (
    <span className="group/tt relative inline-flex">
      <span
        className="inline-flex cursor-help text-muted-foreground hover:text-foreground"
        aria-label={text}
      >
        <HelpCircle className="h-3.5 w-3.5" />
      </span>
      <span
        role="tooltip"
        className="pointer-events-none absolute left-1/2 top-full z-30 mt-1 w-64 -translate-x-1/2 rounded-md border border-border bg-popover px-2 py-1 text-[11px] text-popover-foreground opacity-0 shadow-lg transition-opacity group-hover/tt:opacity-100"
      >
        {text}
      </span>
    </span>
  );
}

function ScopeMultiSelect({
  value,
  onChange,
}: {
  value: string[];
  onChange: (next: string[]) => void;
}) {
  const toggle = (scope: string) => {
    if (value.includes(scope)) {
      onChange(value.filter((s) => s !== scope));
      return;
    }
    onChange([...value, scope]);
  };

  const summary = value.length > 0 ? value.join(", ") : "Select scopes";

  return (
    <details className="relative">
      <summary className="flex h-10 cursor-pointer list-none items-center rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground">
        <span className="truncate">{summary}</span>
      </summary>
      <div className="absolute left-0 top-11 z-20 w-full min-w-[220px] rounded-md border border-border bg-popover p-2 shadow-xl">
        {AVAILABLE_SCOPES.map((scope) => (
          <label key={scope} className="flex cursor-pointer items-center gap-2 rounded px-2 py-1.5 text-sm text-popover-foreground hover:bg-accent">
            <input
              type="checkbox"
              checked={value.includes(scope)}
              onChange={() => toggle(scope)}
            />
            {scope}
          </label>
        ))}
      </div>
    </details>
  );
}

function AuthDeleteConfirmDialog({
  target,
  open,
  onOpenChange,
  onConfirm,
  isDeleting,
}: {
  target: { type: "user" | "project"; id: string; label: string } | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isDeleting: boolean;
}) {
  if (!target) return null;

  const title = target.type === "user" ? "Delete User" : "Delete Project Credential";
  const message =
    target.type === "user"
      ? `Are you sure you want to delete user "${target.label}"? This action cannot be undone.`
      : `Are you sure you want to delete project "${target.label}"? This action cannot be undone.`;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="rounded-full bg-destructive/10 p-2">
              <AlertTriangle className="h-5 w-5 text-destructive" />
            </div>
            <DialogTitle>{title}</DialogTitle>
          </div>
          <DialogDescription className="pt-2">{message}</DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isDeleting}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm} disabled={isDeleting} className="gap-2">
            <Trash2 className="h-4 w-4" />
            {isDeleting ? "Deleting..." : "Delete"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
