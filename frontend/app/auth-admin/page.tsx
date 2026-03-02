"use client";

import { useCallback, useEffect, useState } from "react";
import { AppHeader } from "@/components/AppHeader";
import { AuthAdminPanel } from "@/components/AuthAdminPanel";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { authApi } from "@/lib/api";
import { getErrorMessage } from "@/lib/error";
import { toast } from "@/components/ui/use-toast";

export default function AuthAdminPage() {
  const [isReady, setIsReady] = useState(false);
  const [isAuthed, setIsAuthed] = useState(false);
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("admin123456");
  const [isLoggingIn, setIsLoggingIn] = useState(false);

  useEffect(() => {
    const token =
      typeof window !== "undefined" ? localStorage.getItem("auth_token") : null;
    setIsAuthed(!!token);
    setIsReady(true);
  }, []);

  const handleLogin = useCallback(async () => {
    try {
      setIsLoggingIn(true);
      const result = await authApi.login({ username, password });
      localStorage.setItem("auth_token", result.access_token);
      setIsAuthed(true);
      toast({ title: "Logged in", description: `Welcome ${username}` });
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Unable to sign in.", "login");
      toast({ title: "Login failed", description: msg, variant: "destructive" });
    } finally {
      setIsLoggingIn(false);
    }
  }, [password, username]);

  const handleLogout = useCallback(() => {
    localStorage.removeItem("auth_token");
    setIsAuthed(false);
    toast({ title: "Logged out" });
  }, []);

  return (
    <div className="min-h-screen bg-background">
      {!isReady ? null : !isAuthed ? (
        <main className="min-h-screen flex items-center justify-center p-4">
          <div className="mx-auto w-full max-w-sm space-y-4 rounded-2xl border border-border/80 bg-card/90 p-5 shadow-2xl shadow-black/20">
            <div>
              <h2 className="text-lg font-semibold">CMS Login</h2>
              <p className="text-sm text-muted-foreground">
                Sign in before accessing Auth Admin.
              </p>
            </div>
            <div className="space-y-2">
              <Input
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Username"
              />
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Password"
              />
            </div>
            <Button onClick={handleLogin} disabled={isLoggingIn} className="w-full">
              {isLoggingIn ? "Signing in..." : "Sign in"}
            </Button>
          </div>
        </main>
      ) : (
        <>
          <AppHeader />
          <main className="container mx-auto px-4 py-8 space-y-6">
            <div className="flex justify-end">
              <Button variant="outline" size="sm" className="bg-card/80" onClick={handleLogout}>
                Logout
              </Button>
            </div>
            <AuthAdminPanel />
          </main>
        </>
      )}
    </div>
  );
}
