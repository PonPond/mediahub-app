"use client";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useMediaUsage } from "@/hooks/useMedia";
import { formatDate } from "@/lib/utils";
import { LinkIcon } from "lucide-react";

interface UsageInfoModalProps {
  mediaId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function UsageInfoModal({
  mediaId,
  open,
  onOpenChange,
}: UsageInfoModalProps) {
  const { data: usage, isLoading } = useMediaUsage(mediaId ?? "");
  const references = usage?.references ?? [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="rounded-full bg-primary/10 p-2">
              <LinkIcon className="h-5 w-5 text-primary" />
            </div>
            <div>
              <DialogTitle>File Usage</DialogTitle>
              {usage && (
                <DialogDescription>
                  Referenced by{" "}
                  <strong>{usage.ref_count}</strong>{" "}
                  {usage.ref_count === 1 ? "service" : "services"}
                </DialogDescription>
              )}
            </div>
          </div>
        </DialogHeader>

        <div className="space-y-3 max-h-[400px] overflow-y-auto">
          {isLoading ? (
            Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-16 w-full rounded-md" />
            ))
          ) : references.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <LinkIcon className="h-8 w-8 mb-2 opacity-40" />
              <p className="text-sm">No references found</p>
              <p className="text-xs mt-1">This file is not used anywhere yet.</p>
            </div>
          ) : (
            references.map((ref) => (
              <div
                key={ref.id}
                className="rounded-md border p-3 space-y-2"
              >
                <div className="flex items-center justify-between">
                  <Badge variant="secondary">{ref.ref_service}</Badge>
                  <span className="text-xs text-muted-foreground">
                    {formatDate(ref.created_at)}
                  </span>
                </div>
                <div className="grid grid-cols-3 gap-1 text-xs">
                  <div>
                    <span className="text-muted-foreground">Table: </span>
                    <span className="font-mono">{ref.ref_table}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">ID: </span>
                    <span className="font-mono">{ref.ref_id}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Field: </span>
                    <span className="font-mono">{ref.ref_field}</span>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
