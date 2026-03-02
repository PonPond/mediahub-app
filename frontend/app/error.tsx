"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Application error:", error);
  }, [error]);

  return (
    <div className="min-h-screen flex items-center justify-center p-8">
      <div className="max-w-md w-full space-y-4 text-center">
        <h2 className="text-xl font-semibold">Something went wrong</h2>
        <p className="text-sm text-muted-foreground font-mono bg-muted p-3 rounded text-left break-all">
          {error.message || "An unexpected error occurred"}
        </p>
        {error.digest && (
          <p className="text-xs text-muted-foreground">Digest: {error.digest}</p>
        )}
        <Button onClick={reset}>Try again</Button>
      </div>
    </div>
  );
}
