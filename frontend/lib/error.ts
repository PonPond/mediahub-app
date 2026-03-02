import axios from "axios";

type ErrorContext = "login" | "general";

function normalizeMessage(message: string): string {
  const cleaned = message.trim();
  if (!cleaned) return "Something went wrong. Please try again.";
  return cleaned.charAt(0).toUpperCase() + cleaned.slice(1);
}

export function getErrorMessage(
  err: unknown,
  fallback = "Something went wrong. Please try again.",
  context: ErrorContext = "general"
): string {
  if (axios.isAxiosError(err)) {
    if (!err.response) {
      return "Cannot connect to server. Please try again.";
    }

    const status = err.response.status;
    const data = err.response.data as { error?: unknown } | undefined;
    const apiMessage = typeof data?.error === "string" ? data.error : "";

    if (context === "login" && (status === 401 || apiMessage.toLowerCase().includes("invalid credentials"))) {
      return "Incorrect username or password.";
    }

    if (status === 401) return "Please log in again.";
    if (status === 403) return "You do not have permission for this action.";
    if (status === 404) return "The requested data was not found.";
    if (status === 409) return normalizeMessage(apiMessage || "Cannot complete this action due to a conflict.");
    if (status >= 500) return "Server error. Please try again.";

    if (apiMessage) {
      return normalizeMessage(apiMessage);
    }
  }

  if (err instanceof Error && err.message.trim()) {
    return normalizeMessage(err.message);
  }

  return fallback;
}

