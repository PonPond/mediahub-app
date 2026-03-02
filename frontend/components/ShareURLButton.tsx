"use client";

import { useState } from "react";
import { Share2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { toast } from "@/components/ui/use-toast";

interface ShareURLButtonProps {
  url: string;
  title?: string;
  className?: string;
  size?: "default" | "sm" | "lg" | "icon";
  variant?: "default" | "outline" | "ghost";
}

export function ShareURLButton({
  url,
  title = "Share file",
  className,
  size = "sm",
  variant = "outline",
}: ShareURLButtonProps) {
  const [isSharing, setIsSharing] = useState(false);
  const isIcon = size === "icon";

  const handleShare = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isSharing) return;
    try {
      setIsSharing(true);
      if (navigator.share) {
        await navigator.share({ title, url });
      } else {
        await navigator.clipboard.writeText(url);
        toast({ title: "Link copied", description: "Shared URL copied to clipboard" });
      }
    } catch {
      // Ignore canceled share dialogs and clipboard failures.
    } finally {
      setIsSharing(false);
    }
  };

  return (
    <Button
      variant={variant}
      size={size}
      onClick={handleShare}
      disabled={isSharing}
      title="Share URL"
      aria-label="Share URL"
      className={cn("gap-1.5 transition-all", className)}
    >
      <Share2 className="h-3.5 w-3.5" />
      {!isIcon && <span>Share</span>}
    </Button>
  );
}
