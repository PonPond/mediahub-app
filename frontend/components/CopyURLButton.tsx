"use client";

import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface CopyURLButtonProps {
  url: string;
  className?: string;
  size?: "default" | "sm" | "lg" | "icon";
  variant?: "default" | "outline" | "ghost";
}

export function CopyURLButton({
  url,
  className,
  size = "sm",
  variant = "outline",
}: CopyURLButtonProps) {
  const [copied, setCopied] = useState(false);
  const isIcon = size === "icon";

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for older browsers
      const textarea = document.createElement("textarea");
      textarea.value = url;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand("copy");
      document.body.removeChild(textarea);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <Button
      variant={variant}
      size={size}
      onClick={handleCopy}
      title={copied ? "Copied" : "Copy URL"}
      aria-label={copied ? "Copied" : "Copy URL"}
      className={cn("gap-1.5 transition-all", className)}
    >
      {copied ? <Check className="h-3.5 w-3.5 text-green-500" /> : <Copy className="h-3.5 w-3.5" />}
      {!isIcon && (
        <span className={copied ? "text-green-500" : undefined}>
          {copied ? "Copied!" : "Copy URL"}
        </span>
      )}
    </Button>
  );
}
