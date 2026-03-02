"use client";

import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import { formatBytes, truncateFilename } from "@/lib/utils";
import type { UploadProgress as UploadProgressType } from "@/types/media";

interface UploadProgressProps {
  uploads: UploadProgressType[];
  onClearCompleted: () => void;
  onClearAll: () => void;
}

export function UploadProgress({
  uploads,
  onClearCompleted,
  onClearAll,
}: UploadProgressProps) {
  if (uploads.length === 0) return null;

  const doneCount = uploads.filter((u) => u.status === "done").length;
  const allDone = doneCount === uploads.length;

  return (
    <div className="rounded-lg border bg-card p-4 space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium">
          Uploads ({doneCount}/{uploads.length})
        </h3>
        <div className="flex gap-2">
          {doneCount > 0 && (
            <Button variant="ghost" size="sm" onClick={onClearCompleted}>
              Clear done
            </Button>
          )}
          {allDone && (
            <Button variant="ghost" size="sm" onClick={onClearAll}>
              Clear all
            </Button>
          )}
        </div>
      </div>

      <div className="space-y-2 max-h-64 overflow-y-auto pr-1">
        {uploads.map((upload, i) => (
          <UploadItem key={i} upload={upload} />
        ))}
      </div>
    </div>
  );
}

function UploadItem({ upload }: { upload: UploadProgressType }) {
  return (
    <div className="flex items-center gap-3">
      <StatusIcon status={upload.status} />
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between mb-1">
          <span className="text-xs font-medium truncate">
            {truncateFilename(upload.file.name)}
          </span>
          <span className="text-xs text-muted-foreground ml-2 shrink-0">
            {formatBytes(upload.file.size)}
          </span>
        </div>
        {upload.status === "uploading" && (
          <Progress value={upload.progress} className="h-1.5" />
        )}
        {upload.status === "error" && (
          <p className="text-xs text-destructive">{upload.error}</p>
        )}
        {upload.status === "done" && (
          <p className="text-xs text-green-600">Uploaded successfully</p>
        )}
      </div>
      {upload.status === "uploading" && (
        <span className="text-xs text-muted-foreground shrink-0">
          {upload.progress}%
        </span>
      )}
    </div>
  );
}

function StatusIcon({ status }: { status: UploadProgressType["status"] }) {
  switch (status) {
    case "done":
      return <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0" />;
    case "error":
      return <XCircle className="h-4 w-4 text-destructive shrink-0" />;
    case "uploading":
      return <Loader2 className="h-4 w-4 text-primary shrink-0 animate-spin" />;
    default:
      return <div className="h-4 w-4 rounded-full border-2 border-muted shrink-0" />;
  }
}
