"use client";

import { AlertTriangle, Trash2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import type { MediaFile } from "@/types/media";
import { truncateFilename } from "@/lib/utils";

interface DeleteConfirmDialogProps {
  file: MediaFile | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isDeleting?: boolean;
}

export function DeleteConfirmDialog({
  file,
  open,
  onOpenChange,
  onConfirm,
  isDeleting = false,
}: DeleteConfirmDialogProps) {
  if (!file) return null;

  const isInUse = file.ref_count > 0;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="rounded-full bg-destructive/10 p-2">
              <AlertTriangle className="h-5 w-5 text-destructive" />
            </div>
            <DialogTitle>Delete File</DialogTitle>
          </div>
          <DialogDescription className="pt-2">
            {isInUse ? (
              <span className="text-destructive font-medium">
                This file is referenced by {file.ref_count} service
                {file.ref_count > 1 ? "s" : ""} and cannot be deleted. Remove
                all references first.
              </span>
            ) : (
              <>
                Are you sure you want to delete{" "}
                <strong>{truncateFilename(file.file_name)}</strong>? This action
                cannot be undone — the file will be permanently removed from
                storage.
              </>
            )}
          </DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isDeleting}
          >
            Cancel
          </Button>
          {!isInUse && (
            <Button
              variant="destructive"
              onClick={onConfirm}
              disabled={isDeleting}
              className="gap-2"
            >
              <Trash2 className="h-4 w-4" />
              {isDeleting ? "Deleting…" : "Delete"}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
