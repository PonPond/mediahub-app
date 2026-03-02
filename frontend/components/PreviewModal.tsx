"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CopyURLButton } from "@/components/CopyURLButton";
import { ShareURLButton } from "@/components/ShareURLButton";
import type { MediaFile } from "@/types/media";
import {
  formatBytes,
  formatDate,
  getMimeGroup,
  isPreviewable,
} from "@/lib/utils";
import { Download, ExternalLink, FileIcon, Globe, Lock } from "lucide-react";

interface PreviewModalProps {
  file: MediaFile | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function PreviewModal({ file, open, onOpenChange }: PreviewModalProps) {
  const [imgError, setImgError] = useState(false);
  const [sqlContent, setSqlContent] = useState("");
  const [sqlLoading, setSqlLoading] = useState(false);
  const [sqlError, setSqlError] = useState<string | null>(null);
  const mimeType = file?.mime_type ?? "";
  const fileName = file?.file_name ?? "";
  const url = file?.url ?? "";
  const group = getMimeGroup(mimeType);
  const previewable = isPreviewable(mimeType);
  const isSqlFile = fileName.toLowerCase().endsWith(".sql") || mimeType.includes("sql");

  useEffect(() => {
    if (!open || !isSqlFile || !url) return;

    const controller = new AbortController();
    setSqlLoading(true);
    setSqlError(null);
    setSqlContent("");

    fetch(url, { signal: controller.signal })
      .then(async (res) => {
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        const text = await res.text();
        setSqlContent(text);
      })
      .catch((err: unknown) => {
        if (err instanceof Error && err.name === "AbortError") return;
        setSqlError("Preview unavailable for this SQL file.");
      })
      .finally(() => {
        setSqlLoading(false);
      });

    return () => controller.abort();
  }, [isSqlFile, open, url]);

  if (!file) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl max-h-[90vh] flex flex-col">
        <DialogHeader>
          <div className="flex items-start justify-between gap-4 pr-6">
            <div className="min-w-0">
              <DialogTitle className="truncate text-base">{file.file_name}</DialogTitle>
              <div className="flex items-center gap-2 mt-1 flex-wrap">
                <Badge variant="outline" className="text-xs">{file.mime_type}</Badge>
                <Badge variant="outline" className="text-xs">{formatBytes(file.size)}</Badge>
                {file.is_public ? (
                  <Badge variant="secondary" className="text-xs gap-1">
                    <Globe className="h-3 w-3" /> Public
                  </Badge>
                ) : (
                  <Badge variant="secondary" className="text-xs gap-1">
                    <Lock className="h-3 w-3" /> Private
                  </Badge>
                )}
              </div>
            </div>
          </div>
        </DialogHeader>

        {/* Preview area */}
        <div className="flex-1 overflow-auto rounded-md bg-muted/30 flex items-center justify-center min-h-[200px] max-h-[500px]">
          {isSqlFile ? (
            <div className="h-full w-full overflow-auto p-4">
              {sqlLoading && <p className="text-sm text-muted-foreground">Loading SQL preview...</p>}
              {!sqlLoading && sqlError && <p className="text-sm text-muted-foreground">{sqlError}</p>}
              {!sqlLoading && !sqlError && (
                <pre className="whitespace-pre-wrap break-words rounded-md border bg-background p-3 text-xs font-mono">
                  {sqlContent || "-- empty file --"}
                </pre>
              )}
            </div>
          ) : previewable && url && !imgError ? (
            <FilePreview file={file} url={url} onImgError={() => setImgError(true)} />
          ) : (
            <div className="flex flex-col items-center gap-3 text-muted-foreground py-12">
              <FileIcon className="h-16 w-16 opacity-40" />
              <p className="text-sm">No preview available</p>
            </div>
          )}
        </div>

        {/* Metadata */}
        <div className="grid grid-cols-2 gap-x-6 gap-y-1 text-xs border-t pt-3">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Uploaded</span>
            <span>{formatDate(file.created_at)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">By</span>
            <span>{file.uploaded_by}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">References</span>
            <span>{file.ref_count}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Service</span>
            <span className="font-mono">{file.source_service || "-"}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Module</span>
            <span className="font-mono">{file.source_module || "-"}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Checksum</span>
            <span className="font-mono truncate max-w-[120px]">{file.checksum.slice(0, 12)}…</span>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 border-t pt-3 flex-wrap">
          {url && <ShareURLButton url={url} title={file.file_name} />}
          {url && <CopyURLButton url={url} />}
          {url && (
            <Button variant="outline" size="sm" asChild className="gap-1.5">
              <a href={url} target="_blank" rel="noopener noreferrer">
                <ExternalLink className="h-3.5 w-3.5" />
                Open
              </a>
            </Button>
          )}
          {url && (
            <Button variant="outline" size="sm" asChild className="gap-1.5">
              <a href={url} download={file.file_name}>
                <Download className="h-3.5 w-3.5" />
                Download
              </a>
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

function FilePreview({
  file,
  url,
  onImgError,
}: {
  file: MediaFile;
  url: string;
  onImgError: () => void;
}) {
  const group = getMimeGroup(file.mime_type);

  if (group === "image") {
    return (
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={url}
        alt={file.file_name}
        className="max-w-full max-h-[460px] object-contain rounded"
        onError={onImgError}
      />
    );
  }

  if (group === "video") {
    return (
      <video
        src={url}
        controls
        className="max-w-full max-h-[460px] rounded"
      />
    );
  }

  if (group === "audio") {
    return (
      <audio
        src={url}
        controls
        className="w-full max-w-2xl"
      />
    );
  }

  if (file.mime_type === "application/pdf") {
    return (
      <iframe
        src={url}
        title={file.file_name}
        className="w-full h-[460px] rounded bg-white"
      />
    );
  }

  return null;
}
