"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { MimeGroup } from "@/types/media";

const ALL_TYPES = "__all__";

const TYPES: { label: string; value: string }[] = [
  { label: "All types", value: ALL_TYPES },
  { label: "Images", value: "image" },
  { label: "Videos", value: "video" },
  { label: "Audio", value: "audio" },
  { label: "Documents", value: "document" },
  { label: "Other", value: "other" },
];

const SORT_OPTIONS = [
  { label: "Name A–Z", value: "file_name:asc" },
  { label: "Name Z–A", value: "file_name:desc" },
  { label: "Newest first", value: "created_at:desc" },
  { label: "Oldest first", value: "created_at:asc" },
  { label: "Largest first", value: "size:desc" },
  { label: "Smallest first", value: "size:asc" },
];

interface FilterDropdownProps {
  type: MimeGroup;
  onTypeChange: (v: MimeGroup) => void;
  sortBy: string;
  sortDir: string;
  onSortChange: (sortBy: string, sortDir: string) => void;
}

export function FilterDropdown({
  type,
  onTypeChange,
  sortBy,
  sortDir,
  onSortChange,
}: FilterDropdownProps) {
  const currentSort = `${sortBy}:${sortDir}`;

  const handleSortChange = (value: string) => {
    const [by, dir] = value.split(":");
    onSortChange(by, dir);
  };

  return (
    <div className="flex items-center gap-2">
      <Select
        value={type || ALL_TYPES}
        onValueChange={(v) => onTypeChange((v === ALL_TYPES ? "" : v) as MimeGroup)}
      >
        <SelectTrigger className="w-[140px]">
          <SelectValue placeholder="File type" />
        </SelectTrigger>
        <SelectContent>
          {TYPES.map((t) => (
            <SelectItem key={t.value} value={t.value}>
              {t.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select value={currentSort} onValueChange={handleSortChange}>
        <SelectTrigger className="w-[160px]">
          <SelectValue placeholder="Sort by" />
        </SelectTrigger>
        <SelectContent>
          {SORT_OPTIONS.map((o) => (
            <SelectItem key={o.value} value={o.value}>
              {o.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
