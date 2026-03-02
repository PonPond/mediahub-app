"use client";

import { Skeleton } from "@/components/ui/skeleton";
import { MediaCard } from "@/components/MediaCard";
import type { MediaFile } from "@/types/media";
import { FolderOpen } from "lucide-react";

interface MediaGridProps {
  files: MediaFile[];
  view: "grid" | "list";
  isLoading: boolean;
  onPreview: (file: MediaFile) => void;
  onDelete: (file: MediaFile) => void;
  onUsage: (file: MediaFile) => void;
}

export function MediaGrid({
  files,
  view,
  isLoading,
  onPreview,
  onDelete,
  onUsage,
}: MediaGridProps) {
  if (isLoading) {
    return (
      <div
        className={
          view === "grid"
            ? "grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4"
            : "flex flex-col gap-2"
        }
      >
        {Array.from({ length: 12 }).map((_, i) => (
          <Skeleton
            key={i}
            className={
              view === "grid"
                ? "aspect-video rounded-lg"
                : "h-16 rounded-lg"
            }
          />
        ))}
      </div>
    );
  }

  if (files.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3">
        <FolderOpen className="h-16 w-16 opacity-30" />
        <p className="text-lg font-medium">No files found</p>
        <p className="text-sm">Upload files or adjust your filters.</p>
      </div>
    );
  }

  if (view === "list") {
    return (
      <div className="flex flex-col gap-2">
        {files.map((file) => (
          <MediaCard
            key={file.id}
            file={file}
            view="list"
            onPreview={onPreview}
            onDelete={onDelete}
            onUsage={onUsage}
          />
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
      {files.map((file) => (
        <MediaCard
          key={file.id}
          file={file}
          view="grid"
          onPreview={onPreview}
          onDelete={onDelete}
          onUsage={onUsage}
        />
      ))}
    </div>
  );
}
