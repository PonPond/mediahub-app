"use client";

import { useState, useCallback, useEffect } from "react";
import { LayoutGrid, List, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FileUploader } from "@/components/FileUploader";
import { UploadProgress } from "@/components/UploadProgress";
import { MediaGrid } from "@/components/MediaGrid";
import { SearchBar } from "@/components/SearchBar";
import { FilterDropdown } from "@/components/FilterDropdown";
import { Pagination } from "@/components/Pagination";
import { PreviewModal } from "@/components/PreviewModal";
import { DeleteConfirmDialog } from "@/components/DeleteConfirmDialog";
import { UsageInfoModal } from "@/components/UsageInfoModal";
import { useMediaList, useDeleteMedia, useFileUpload } from "@/hooks/useMedia";
import { toast } from "@/components/ui/use-toast";
import type { MediaFile, MimeGroup } from "@/types/media";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { mediaKeys } from "@/hooks/useMedia";
import { authApi, mediaApi } from "@/lib/api";
import { getErrorMessage } from "@/lib/error";
import { Input } from "@/components/ui/input";
import { AppHeader } from "@/components/AppHeader";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const ALL_OPTION = "__all__";

export function MediaLibrary() {
  const qc = useQueryClient();
  const [isReady, setIsReady] = useState(false);
  const [isAuthed, setIsAuthed] = useState(false);
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("admin123456");
  const [isLoggingIn, setIsLoggingIn] = useState(false);

  // Filters / Pagination state
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [uploadedBy, setUploadedBy] = useState("");
  const [sourceService, setSourceService] = useState("");
  const [sourceModule, setSourceModule] = useState("");
  const [type, setType] = useState<MimeGroup>("");
  const [sortBy, setSortBy] = useState("file_name");
  const [sortDir, setSortDir] = useState("asc");
  const [view, setView] = useState<"grid" | "list">("grid");

  // Modal state
  const [previewFile, setPreviewFile] = useState<MediaFile | null>(null);
  const [deleteFile, setDeleteFile] = useState<MediaFile | null>(null);
  const [usageMediaId, setUsageMediaId] = useState<string | null>(null);

  // Data
  const { data, isLoading, refetch } = useMediaList({
    page,
    limit: 20,
    type: type || undefined,
    search: search || undefined,
    uploaded_by: uploadedBy || undefined,
    source_service: sourceService || undefined,
    source_module: sourceModule || undefined,
    sort_by: sortBy,
    sort_dir: sortDir,
  });
  const { data: filterOptions } = useQuery({
    queryKey: ["media", "filter-options"],
    queryFn: () => mediaApi.getFilterOptions(),
    staleTime: 60_000,
  });

  const deleteMutation = useDeleteMedia();

  const { uploads, uploadFiles, clearCompleted, clearAll } = useFileUpload({
    onAllDone: () => {
      qc.invalidateQueries({ queryKey: mediaKeys.lists() });
    },
  });

  // Handlers
  const handleSearch = useCallback((v: string) => {
    setSearch(v);
    setPage(1);
  }, []);

  const handleTypeChange = useCallback((v: MimeGroup) => {
    setType(v);
    setPage(1);
  }, []);

  const handleSourceServiceChange = useCallback((v: string) => {
    setSourceService(v);
    setPage(1);
  }, []);

  const handleUploadedByChange = useCallback((v: string) => {
    setUploadedBy(v);
    setPage(1);
  }, []);

  const handleSourceModuleChange = useCallback((v: string) => {
    setSourceModule(v);
    setPage(1);
  }, []);

  const handleSortChange = useCallback((by: string, dir: string) => {
    setSortBy(by);
    setSortDir(dir);
    setPage(1);
  }, []);

  const handleDelete = useCallback(async () => {
    if (!deleteFile) return;
    try {
      await deleteMutation.mutateAsync(deleteFile.id);
      toast({ title: "File deleted", description: deleteFile.file_name });
      setDeleteFile(null);
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to delete file.");
      toast({ title: "Delete failed", description: msg, variant: "destructive" });
    }
  }, [deleteFile, deleteMutation]);

  const files = data?.items ?? [];

  useEffect(() => {
    const token =
      typeof window !== "undefined" ? localStorage.getItem("auth_token") : null;
    setIsAuthed(!!token);
    setIsReady(true);
  }, []);

  const handleLogin = useCallback(async () => {
    try {
      setIsLoggingIn(true);
      const result = await authApi.login({ username, password });
      localStorage.setItem("auth_token", result.access_token);
      setIsAuthed(true);
      toast({ title: "Logged in", description: `Welcome ${username}` });
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Unable to sign in.", "login");
      toast({ title: "Login failed", description: msg, variant: "destructive" });
    } finally {
      setIsLoggingIn(false);
    }
  }, [password, username]);

  const handleLogout = useCallback(() => {
    localStorage.removeItem("auth_token");
    setIsAuthed(false);
    toast({ title: "Logged out" });
  }, []);

  if (!isReady) {
    return null;
  }

  if (!isAuthed) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <div className="mx-auto w-full max-w-sm space-y-4 rounded-2xl border border-border/80 bg-card/90 p-5 shadow-2xl shadow-black/20">
          <div>
            <h2 className="text-lg font-semibold">CMS Login</h2>
            <p className="text-sm text-muted-foreground">
              Sign in before accessing the media library.
            </p>
          </div>
          <div className="space-y-2">
            <Input
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Username"
            />
            <Input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Password"
            />
          </div>
          <Button onClick={handleLogin} disabled={isLoggingIn} className="w-full">
            {isLoggingIn ? "Signing in..." : "Sign in"}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <AppHeader />
      <main className="container mx-auto px-4 py-8">
        <div className="space-y-6">
          <div className="flex justify-end">
            <Button variant="outline" size="sm" className="bg-card/80" onClick={handleLogout}>
              Logout
            </Button>
          </div>

          {/* Upload area */}
          <section className="rounded-2xl border border-border/80 bg-card/80 p-4">
            <FileUploader onFiles={uploadFiles} />
            <UploadProgress
              uploads={uploads}
              onClearCompleted={clearCompleted}
              onClearAll={clearAll}
            />
          </section>

          {/* Controls */}
          <section className="flex flex-col items-start gap-3 rounded-2xl border border-border/80 bg-card/80 p-4 sm:flex-row sm:items-center">
            <SearchBar
              value={search}
              onChange={handleSearch}
              placeholder="Filter name..."
              className="w-full sm:w-72"
            />
            <SearchBar
              value={uploadedBy}
              onChange={handleUploadedByChange}
              placeholder="Filter by..."
              className="w-full sm:w-56"
            />
            <Select
              value={sourceService || ALL_OPTION}
              onValueChange={(v) => handleSourceServiceChange(v === ALL_OPTION ? "" : v)}
            >
              <SelectTrigger className="w-full sm:w-56">
                <SelectValue placeholder="Filter service" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL_OPTION}>All services</SelectItem>
                {(filterOptions?.source_services ?? []).map((service) => (
                  <SelectItem key={service} value={service}>
                    {service}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={sourceModule || ALL_OPTION}
              onValueChange={(v) => handleSourceModuleChange(v === ALL_OPTION ? "" : v)}
            >
              <SelectTrigger className="w-full sm:w-56">
                <SelectValue placeholder="Filter module" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL_OPTION}>All modules</SelectItem>
                {(filterOptions?.source_modules ?? []).map((module) => (
                  <SelectItem key={module} value={module}>
                    {module}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <FilterDropdown
              type={type}
              onTypeChange={handleTypeChange}
              sortBy={sortBy}
              sortDir={sortDir}
              onSortChange={handleSortChange}
            />
            <div className="ml-auto flex items-center gap-2">
              <Button variant="outline" size="icon" onClick={() => refetch()} title="Refresh">
                <RefreshCw className="h-4 w-4" />
              </Button>
              <div className="flex rounded-md border">
                <Button
                  variant={view === "grid" ? "default" : "ghost"}
                  size="icon"
                  className="rounded-r-none border-0"
                  onClick={() => setView("grid")}
                >
                  <LayoutGrid className="h-4 w-4" />
                </Button>
                <Button
                  variant={view === "list" ? "default" : "ghost"}
                  size="icon"
                  className="rounded-l-none border-0"
                  onClick={() => setView("list")}
                >
                  <List className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </section>

          {/* Stats */}
          {data && (
            <p className="text-sm text-muted-foreground">
              {data.total.toLocaleString()} file{data.total !== 1 ? "s" : ""}
              {search && ` matching name "${search}"`}
            </p>
          )}

          {/* Grid / List */}
          <section className="rounded-2xl border border-border/80 bg-card/70 p-3 sm:p-4">
            <MediaGrid
              files={files}
              view={view}
              isLoading={isLoading}
              onPreview={setPreviewFile}
              onDelete={setDeleteFile}
              onUsage={(f) => setUsageMediaId(f.id)}
            />
          </section>

          {/* Pagination */}
          {data && data.total_pages > 1 && (
            <Pagination
              page={page}
              totalPages={data.total_pages}
              onPageChange={setPage}
              className="pt-4"
            />
          )}

          {/* Modals */}
          <PreviewModal
            file={previewFile}
            open={!!previewFile}
            onOpenChange={(o) => !o && setPreviewFile(null)}
          />
          <DeleteConfirmDialog
            file={deleteFile}
            open={!!deleteFile}
            onOpenChange={(o) => !o && setDeleteFile(null)}
            onConfirm={handleDelete}
            isDeleting={deleteMutation.isPending}
          />
          <UsageInfoModal
            mediaId={usageMediaId}
            open={!!usageMediaId}
            onOpenChange={(o) => !o && setUsageMediaId(null)}
          />
        </div>
      </main>
    </div>
  );
}
