"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { mediaApi, referenceApi } from "@/lib/api";
import { getErrorMessage } from "@/lib/error";
import type {
  AddReferenceInput,
  ListParams,
  RemoveReferenceInput,
  UploadProgress,
} from "@/types/media";
import { useCallback, useRef, useState } from "react";

// ─── Query Keys ───────────────────────────────────────────────────────────────

export const mediaKeys = {
  all: ["media"] as const,
  lists: () => [...mediaKeys.all, "list"] as const,
  list: (params: ListParams) => [...mediaKeys.lists(), params] as const,
  detail: (id: string) => [...mediaKeys.all, "detail", id] as const,
  usage: (id: string) => [...mediaKeys.all, "usage", id] as const,
};

// ─── Hooks ────────────────────────────────────────────────────────────────────

export function useMediaList(params: ListParams = {}) {
  return useQuery({
    queryKey: mediaKeys.list(params),
    queryFn: () => mediaApi.list(params),
    staleTime: 30_000,
  });
}

export function useMediaDetail(id: string) {
  return useQuery({
    queryKey: mediaKeys.detail(id),
    queryFn: () => mediaApi.getById(id),
    enabled: !!id,
  });
}

export function useMediaUsage(id: string) {
  return useQuery({
    queryKey: mediaKeys.usage(id),
    queryFn: () => mediaApi.getUsage(id),
    enabled: !!id,
  });
}

export function useDeleteMedia() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => mediaApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: mediaKeys.lists() });
    },
  });
}

export function useAddReference() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AddReferenceInput) => referenceApi.add(input),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: mediaKeys.usage(variables.media_id) });
      qc.invalidateQueries({ queryKey: mediaKeys.lists() });
    },
  });
}

export function useRemoveReference() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: RemoveReferenceInput) => referenceApi.remove(input),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: mediaKeys.usage(variables.media_id) });
      qc.invalidateQueries({ queryKey: mediaKeys.lists() });
    },
  });
}

// ─── Multi-file upload with progress ─────────────────────────────────────────

export function useFileUpload(opts: {
  sourceService?: string;
  sourceModule?: string;
  isPublic?: boolean;
  onAllDone?: () => void;
}) {
  const qc = useQueryClient();
  const [uploads, setUploads] = useState<UploadProgress[]>([]);
  const abortRefs = useRef<Map<string, AbortController>>(new Map());

  const updateUpload = useCallback(
    (file: File, patch: Partial<UploadProgress>) => {
      setUploads((prev) =>
        prev.map((u) => (u.file === file ? { ...u, ...patch } : u))
      );
    },
    []
  );

  const uploadFiles = useCallback(
    async (files: File[]) => {
      const newUploads: UploadProgress[] = files.map((f) => ({
        file: f,
        progress: 0,
        status: "pending",
      }));
      setUploads((prev) => [...prev, ...newUploads]);

      await Promise.allSettled(
        files.map(async (file) => {
          updateUpload(file, { status: "uploading" });
          try {
            const result = await mediaApi.upload(file, {
              ...opts,
              onProgress: (pct) => updateUpload(file, { progress: pct }),
            });
            updateUpload(file, { status: "done", progress: 100, result });
          } catch (err: unknown) {
            const msg = getErrorMessage(err, "Upload failed.");
            updateUpload(file, { status: "error", error: msg });
          }
        })
      );

      qc.invalidateQueries({ queryKey: mediaKeys.lists() });
      opts.onAllDone?.();
    },
    [opts, qc, updateUpload]
  );

  const clearCompleted = useCallback(() => {
    setUploads((prev) => prev.filter((u) => u.status !== "done"));
  }, []);

  const clearAll = useCallback(() => {
    setUploads([]);
  }, []);

  return { uploads, uploadFiles, clearCompleted, clearAll };
}
