"use client";

import { useEffect, useRef, useState } from "react";
import { Search, X } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  debounceMs?: number;
}

export function SearchBar({
  value,
  onChange,
  placeholder = "Search files…",
  className,
  debounceMs = 350,
}: SearchBarProps) {
  const [local, setLocal] = useState(value);
  const timer = useRef<ReturnType<typeof setTimeout>>();

  // Sync external value changes (e.g. clear)
  useEffect(() => {
    setLocal(value);
  }, [value]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const v = e.target.value;
    setLocal(v);
    clearTimeout(timer.current);
    timer.current = setTimeout(() => onChange(v), debounceMs);
  };

  const handleClear = () => {
    setLocal("");
    clearTimeout(timer.current);
    onChange("");
  };

  return (
    <div className={cn("relative flex items-center", className)}>
      <Search className="absolute left-3 h-4 w-4 text-muted-foreground pointer-events-none" />
      <Input
        value={local}
        onChange={handleChange}
        placeholder={placeholder}
        className="pl-9 pr-9"
      />
      {local && (
        <Button
          variant="ghost"
          size="icon"
          className="absolute right-1 h-7 w-7"
          onClick={handleClear}
        >
          <X className="h-3.5 w-3.5" />
        </Button>
      )}
    </div>
  );
}
