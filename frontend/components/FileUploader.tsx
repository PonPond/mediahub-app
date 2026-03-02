"use client";

import { useCallback } from "react";
import { useDropzone } from "react-dropzone";
import { Upload, UploadCloud } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

interface FileUploaderProps {
  onFiles: (files: File[]) => void;
  isPublic?: boolean;
  disabled?: boolean;
  className?: string;
  maxSizeBytes?: number;
}

export function FileUploader({
  onFiles,
  disabled = false,
  className,
  maxSizeBytes = 100 * 1024 * 1024, // 100 MB
}: FileUploaderProps) {
  const onDrop = useCallback(
    (accepted: File[]) => {
      if (accepted.length > 0) onFiles(accepted);
    },
    [onFiles]
  );

  const { getRootProps, getInputProps, isDragActive, isDragReject, fileRejections } =
    useDropzone({
      onDrop,
      maxSize: maxSizeBytes,
      disabled,
      multiple: true,
    });

  return (
    <div
      {...getRootProps()}
      className={cn(
        "flex flex-col items-center justify-center gap-3 rounded-xl border-2 border-dashed p-8 cursor-pointer transition-all duration-200 select-none",
        isDragActive && !isDragReject && "border-primary bg-primary/5 scale-[1.01]",
        isDragReject && "border-destructive bg-destructive/5",
        !isDragActive && !disabled && "border-muted-foreground/25 hover:border-primary/50 hover:bg-muted/30",
        disabled && "opacity-50 cursor-not-allowed",
        className
      )}
    >
      <input {...getInputProps()} />

      <div
        className={cn(
          "rounded-full p-4 transition-colors",
          isDragActive && !isDragReject
            ? "bg-primary/10 text-primary"
            : "bg-muted text-muted-foreground"
        )}
      >
        {isDragActive ? (
          <UploadCloud className="h-8 w-8" />
        ) : (
          <Upload className="h-8 w-8" />
        )}
      </div>

      <div className="text-center">
        {isDragReject ? (
          <p className="font-medium text-destructive">File type not supported</p>
        ) : isDragActive ? (
          <p className="font-medium text-primary">Drop to upload</p>
        ) : (
          <>
            <p className="font-medium">Drag & drop files here</p>
            <p className="text-sm text-muted-foreground mt-1">
              or{" "}
              <Button variant="link" className="p-0 h-auto text-sm font-medium">
                browse
              </Button>{" "}
              to choose files
            </p>
          </>
        )}
      </div>

      <p className="text-xs text-muted-foreground">
        All file types — max {Math.round(maxSizeBytes / 1024 / 1024)} MB each
      </p>

      {fileRejections.length > 0 && (
        <div className="w-full rounded-md bg-destructive/10 p-2">
          {fileRejections.map(({ file, errors }) => (
            <p key={file.name} className="text-xs text-destructive">
              {file.name}: {errors.map((e) => e.message).join(", ")}
            </p>
          ))}
        </div>
      )}
    </div>
  );
}
