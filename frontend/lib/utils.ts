import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatBytes(bytes: number, decimals = 2): string {
  if (!Number.isFinite(bytes) || bytes < 0) return "Unknown size";
  if (bytes === 0) return "0 B";
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
}

export function formatDate(dateStr: string): string {
  return new Intl.DateTimeFormat("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(dateStr));
}

export function getMimeGroup(mimeType: string): string {
  if (mimeType.startsWith("image/")) return "image";
  if (mimeType.startsWith("video/")) return "video";
  if (mimeType.startsWith("audio/")) return "audio";
  if (
    mimeType === "application/pdf" ||
    mimeType.startsWith("text/") ||
    mimeType.includes("document") ||
    mimeType.includes("spreadsheet") ||
    mimeType.includes("presentation")
  )
    return "document";
  return "other";
}

export function getMimeIcon(mimeType: string): string {
  const group = getMimeGroup(mimeType);
  switch (group) {
    case "image":
      return "image";
    case "video":
      return "video";
    case "audio":
      return "music";
    case "document":
      return "file-text";
    default:
      return "file";
  }
}

export function isPreviewable(mimeType: string): boolean {
  return (
    mimeType.startsWith("image/") ||
    mimeType.startsWith("video/") ||
    mimeType.startsWith("audio/") ||
    mimeType === "application/pdf"
  );
}

export type FileKind =
  | "image"
  | "video"
  | "audio"
  | "pdf"
  | "word"
  | "excel"
  | "powerpoint"
  | "archive"
  | "code"
  | "text"
  | "file";

export function getFileKind(fileName: string, mimeType: string): FileKind {
  const ext = fileName.toLowerCase().split(".").pop() ?? "";
  if (mimeType.startsWith("image/")) return "image";
  if (mimeType.startsWith("video/")) return "video";
  if (mimeType.startsWith("audio/")) return "audio";
  if (mimeType === "application/pdf" || ext === "pdf") return "pdf";
  if (["doc", "docx", "odt", "rtf"].includes(ext)) return "word";
  if (["xls", "xlsx", "csv", "ods"].includes(ext)) return "excel";
  if (["ppt", "pptx", "odp", "key"].includes(ext)) return "powerpoint";
  if (["zip", "rar", "7z", "tar", "gz"].includes(ext)) return "archive";
  if (
    ["js", "ts", "tsx", "jsx", "go", "java", "py", "php", "rb", "rs", "c", "cpp", "h", "hpp"].includes(ext) ||
    mimeType.includes("json") ||
    mimeType.includes("javascript")
  ) {
    return "code";
  }
  if (mimeType.startsWith("text/")) return "text";
  return "file";
}

export function truncateFilename(name: string, maxLen = 28): string {
  if (name.length <= maxLen) return name;
  const ext = name.lastIndexOf(".");
  if (ext === -1) return name.slice(0, maxLen) + "…";
  const extPart = name.slice(ext);
  const basePart = name.slice(0, maxLen - extPart.length - 1);
  return `${basePart}…${extPart}`;
}
