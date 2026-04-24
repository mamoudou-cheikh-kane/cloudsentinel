"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import { scenariosApi } from "@/lib/api";
import type { Scenario } from "@/lib/api";

const POLL_INTERVAL_MS = 5000;

export interface UseScenariosResult {
  scenarios: Scenario[];
  isLoading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
}

export function useScenarios(): UseScenariosResult {
  const [scenarios, setScenarios] = useState<Scenario[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const pollTimer = useRef<NodeJS.Timeout | null>(null);

  const refresh = useCallback(async () => {
    try {
      const data = await scenariosApi.list();
      setScenarios(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
    pollTimer.current = setInterval(() => {
      void refresh();
    }, POLL_INTERVAL_MS);
    return () => {
      if (pollTimer.current) {
        clearInterval(pollTimer.current);
      }
    };
  }, [refresh]);

  return { scenarios, isLoading, error, refresh };
}
