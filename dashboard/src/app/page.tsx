"use client";

import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { CreateScenarioDialog } from "@/components/create-scenario-dialog";
import { RunButton } from "@/components/run-button";
import { StatusBadge } from "@/components/status-badge";
import { useScenarios } from "@/hooks/use-scenarios";

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

export default function HomePage() {
  const { scenarios, isLoading, error, refresh } = useScenarios();

  return (
    <main className="container mx-auto max-w-6xl p-8">
      <header className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">CloudSentinel</h1>
          <p className="text-muted-foreground mt-1">
            Chaos engineering scenarios for Kubernetes.
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => void refresh()}>
            Refresh
          </Button>
          <CreateScenarioDialog onCreated={() => void refresh()} />
        </div>
      </header>

      {error ? (
        <div className="border-destructive bg-destructive/10 text-destructive mb-6 rounded-md border p-4">
          <strong>Error:</strong> {error}
        </div>
      ) : null}

      <section className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-16">ID</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Updated</TableHead>
              <TableHead className="w-28 text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && scenarios.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-muted-foreground py-8 text-center">
                  Loading scenarios…
                </TableCell>
              </TableRow>
            ) : scenarios.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-muted-foreground py-8 text-center">
                  No scenarios yet. Click “+ New Scenario” to create one.
                </TableCell>
              </TableRow>
            ) : (
              scenarios.map((s) => (
                <TableRow key={s.id}>
                  <TableCell className="font-mono">{s.id}</TableCell>
                  <TableCell className="font-medium">{s.name}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {s.description || "—"}
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={s.status} />
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs">
                    {formatDate(s.created_at)}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs">
                    {formatDate(s.updated_at)}
                  </TableCell>
                  <TableCell className="text-right">
                    <RunButton
                      scenarioId={s.id}
                      status={s.status}
                      onCompleted={() => void refresh()}
                    />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </section>
    </main>
  );
}
