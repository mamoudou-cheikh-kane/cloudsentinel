import { Badge } from "@/components/ui/badge";
import type { ScenarioStatus } from "@/lib/api";

interface StatusBadgeProps {
  status: ScenarioStatus;
}

const STATUS_VARIANT: Record<ScenarioStatus, "default" | "secondary" | "destructive" | "outline"> = {
  pending: "secondary",
  running: "default",
  completed: "outline",
  failed: "destructive",
};

const STATUS_LABEL: Record<ScenarioStatus, string> = {
  pending: "Pending",
  running: "Running",
  completed: "Completed",
  failed: "Failed",
};

export function StatusBadge({ status }: StatusBadgeProps) {
  return <Badge variant={STATUS_VARIANT[status]}>{STATUS_LABEL[status]}</Badge>;
}
