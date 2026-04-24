"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { scenariosApi } from "@/lib/api";
import type { ScenarioStatus } from "@/lib/api";

interface RunButtonProps {
  scenarioId: number;
  status: ScenarioStatus;
  onCompleted: () => void;
}

export function RunButton({ scenarioId, status, onCompleted }: RunButtonProps) {
  const [isRunning, setIsRunning] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleRun = async () => {
    setError(null);
    setIsRunning(true);
    try {
      await scenariosApi.run(scenarioId);
      onCompleted();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Run failed");
    } finally {
      setIsRunning(false);
    }
  };

  const disabled = isRunning || status === "running";

  return (
    <div className="flex flex-col items-end gap-1">
      <Button
        size="sm"
        variant="secondary"
        disabled={disabled}
        onClick={() => void handleRun()}
      >
        {isRunning ? "Running…" : "Run"}
      </Button>
      {error ? (
        <span className="text-destructive text-xs">{error}</span>
      ) : null}
    </div>
  );
}
