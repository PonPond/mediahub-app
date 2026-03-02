"use client";

import { useState } from "react";
import {
  Eye,
  FileArchive,
  FileCode2,
  FileSpreadsheet,
  FileText,
  FileType,
  Globe,
  LinkIcon,
  Lock,
  Music,
  Trash2,
  Video,
  ImageIcon,
  File,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CopyURLButton } from "@/components/CopyURLButton";
import { ShareURLButton } from "@/components/ShareURLButton";
import type { MediaFile } from "@/types/media";
import { formatBytes, formatDate, getFileKind, getMimeGroup, truncateFilename } from "@/lib/utils";
import { cn } from "@/lib/utils";

interface MediaCardProps {
  file: MediaFile;
  view: "grid" | "list";
  onPreview: (file: MediaFile) => void;
  onDelete: (file: MediaFile) => void;
  onUsage: (file: MediaFile) => void;
}

export function MediaCard({
  file,
  view,
  onPreview,
  onDelete,
  onUsage,
}: MediaCardProps) {
  const [imgError, setImgError] = useState(false);
  const group = getMimeGroup(file.mime_type);
  const isImage = group === "image";
  const showThumb = isImage && file.url && !imgError;

  if (view === "list") {
    return (
      <div className="flex items-center gap-4 rounded-lg border bg-card px-4 py-3 hover:bg-muted/30 transition-colors group">
        {/* Thumbnail / Icon */}
        <div className="h-10 w-10 shrink-0 rounded overflow-hidden bg-muted flex items-center justify-center">
          {showThumb ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={file.url}
              alt={file.file_name}
              className="h-full w-full object-cover"
              onError={() => setImgError(true)}
            />
          ) : (
            <MimeIcon group={group} file={file} className="h-5 w-5 text-muted-foreground" />
          )}
        </div>

        {/* Name + meta */}
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium truncate">{truncateFilename(file.file_name, 50)}</p>
          <p className="text-xs text-muted-foreground">
            {formatBytes(file.size)} · {formatDate(file.created_at)}
          </p>
          <p className="text-xs text-muted-foreground truncate">
            By {file.uploaded_by || "-"}
          </p>
        </div>

        {/* Badges */}
        <div className="hidden sm:flex items-center gap-2">
          <Badge variant="outline" className="text-xs capitalize">{group}</Badge>
          {file.is_public ? (
            <Globe className="h-3.5 w-3.5 text-muted-foreground" />
          ) : (
            <Lock className="h-3.5 w-3.5 text-muted-foreground" />
          )}
          {file.ref_count > 0 && (
            <Badge variant="secondary" className="text-xs gap-1">
              <LinkIcon className="h-3 w-3" />
              {file.ref_count}
            </Badge>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity">
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => onPreview(file)}>
            <Eye className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => onUsage(file)}>
            <LinkIcon className="h-4 w-4" />
          </Button>
          {file.url && (
            <ShareURLButton
              url={file.url}
              title={file.file_name}
              size="icon"
              variant="ghost"
              className="h-8 w-8"
            />
          )}
          {file.url && (
            <CopyURLButton
              url={file.url}
              size="icon"
              variant="ghost"
              className="h-8 w-8"
            />
          )}
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-destructive hover:text-destructive"
            onClick={() => onDelete(file)}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>
    );
  }

  // Grid view
  return (
    <div className="group relative rounded-lg border bg-card overflow-hidden hover:shadow-md transition-all">
      {/* Thumbnail */}
      <div
        className="aspect-video bg-muted cursor-pointer relative overflow-hidden"
        onClick={() => onPreview(file)}
      >
        {showThumb ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={file.url}
            alt={file.file_name}
            className="h-full w-full object-cover transition-transform group-hover:scale-105"
            onError={() => setImgError(true)}
          />
        ) : (
          <div className="flex h-full items-center justify-center">
            <MimeIcon group={group} file={file} className="h-12 w-12 text-muted-foreground/50" />
          </div>
        )}

        {/* Overlay */}
        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center opacity-0 group-hover:opacity-100">
          <Eye className="h-6 w-6 text-white" />
        </div>

        {/* Badges overlaid */}
        <div className="absolute top-2 right-2 flex gap-1">
          {file.is_public ? (
            <span className="rounded-full bg-black/50 p-1">
              <Globe className="h-3 w-3 text-white" />
            </span>
          ) : (
            <span className="rounded-full bg-black/50 p-1">
              <Lock className="h-3 w-3 text-white" />
            </span>
          )}
        </div>
      </div>

      {/* Info */}
      <div className="p-3 space-y-1">
        <p className="text-sm font-medium leading-tight truncate">
          {truncateFilename(file.file_name)}
        </p>
        <p className="text-xs text-muted-foreground truncate">
          By {file.uploaded_by || "-"}
        </p>
        <div className="flex items-center justify-between">
          <span className="text-xs text-muted-foreground">{formatBytes(file.size)}</span>
          <span className="text-xs text-muted-foreground">{formatDate(file.created_at).split(",")[0]}</span>
        </div>
        {file.ref_count > 0 && (
          <Badge variant="secondary" className="text-xs gap-1 w-fit">
            <LinkIcon className="h-3 w-3" />
            {file.ref_count} {file.ref_count === 1 ? "reference" : "references"}
          </Badge>
        )}
      </div>

      {/* Actions bar */}
      <div className="flex items-center gap-1 px-3 pb-3 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity">
        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => onUsage(file)}>
          <LinkIcon className="h-3.5 w-3.5" />
        </Button>
        {file.url && (
          <ShareURLButton
            url={file.url}
            title={file.file_name}
            size="icon"
            variant="ghost"
            className="h-7 w-7"
          />
        )}
        {file.url && (
          <CopyURLButton
            url={file.url}
            size="icon"
            variant="ghost"
            className="h-7 w-7"
          />
        )}
        <div className="flex-1" />
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7 text-destructive hover:text-destructive"
          onClick={() => onDelete(file)}
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  );
}

function MimeIcon({
  group,
  className,
  file,
}: {
  group: string;
  className?: string;
  file?: MediaFile;
}) {
  const kind = file ? getFileKind(file.file_name, file.mime_type) : group;
  switch (kind) {
    case "image":
      return <ImageIcon className={cn(className)} />;
    case "video":
      return <Video className={cn(className)} />;
    case "audio":
      return <Music className={cn(className)} />;
    case "pdf":
      return <FileText className={cn(className)} />;
    case "word":
      return <FileType className={cn(className)} />;
    case "excel":
      return <FileSpreadsheet className={cn(className)} />;
    case "powerpoint":
      return <FileType className={cn(className)} />;
    case "archive":
      return <FileArchive className={cn(className)} />;
    case "code":
      return <FileCode2 className={cn(className)} />;
    case "text":
      return <FileText className={cn(className)} />;
    default:
      return group === "document" ? <FileText className={cn(className)} /> : <File className={cn(className)} />;
  }
}
