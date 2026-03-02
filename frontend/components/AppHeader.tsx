"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { HardDrive, Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { cn } from "@/lib/utils";

export function AppHeader() {
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();

  return (
    <header className="sticky top-0 z-20 border-b border-border/80 bg-background/90 backdrop-blur-xl">
      <div className="container mx-auto flex h-16 items-center justify-between gap-3 px-4">
        <div className="flex items-center gap-3">
          <div className="rounded-lg bg-primary/15 p-1.5 ring-1 ring-primary/30">
            <HardDrive className="h-5 w-5 text-primary" />
          </div>
          <div>
            <h1 className="text-lg font-semibold leading-tight">MediaHub</h1>
            <p className="text-xs text-muted-foreground">Storage, Auth and Project Credentials</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <nav className="flex items-center gap-2 rounded-xl border border-border/80 bg-card/80 p-1">
            <Link
              href="/"
              className={cn(
                "rounded-lg px-3 py-1.5 text-sm",
                pathname === "/" ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-muted"
              )}
            >
              Media
            </Link>
            <Link
              href="/auth-admin"
              className={cn(
                "rounded-lg px-3 py-1.5 text-sm",
                pathname === "/auth-admin"
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:bg-muted"
              )}
            >
              Auth Admin
            </Link>
          </nav>

          <button
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            className="rounded-lg border border-border/80 bg-card/80 p-2 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            aria-label="Toggle theme"
          >
            {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
          </button>
        </div>
      </div>
    </header>
  );
}
