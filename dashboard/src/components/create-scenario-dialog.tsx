"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { scenariosApi, FAULT_TYPES } from "@/lib/api";
import type { FaultType } from "@/lib/api";

interface CreateScenarioDialogProps {
  onCreated: () => void;
}

export function CreateScenarioDialog({ onCreated }: CreateScenarioDialogProps) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [faultType, setFaultType] = useState<FaultType>("cpu_stress");
  const [duration, setDuration] = useState(30);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const reset = () => {
    setName("");
    setDescription("");
    setFaultType("cpu_stress");
    setDuration(30);
    setError(null);
  };

  const handleSubmit = async () => {
    setError(null);
    setIsSubmitting(true);
    try {
      await scenariosApi.create({
        name,
        description,
        faults: [{ type: faultType, duration_seconds: duration }],
      });
      reset();
      setOpen(false);
      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create scenario");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        setOpen(v);
        if (!v) reset();
      }}
    >
      <DialogTrigger asChild>
        <Button>+ New Scenario</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Create a new chaos scenario</DialogTitle>
          <DialogDescription>
            Define a fault to inject into the cluster.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              placeholder="e.g. cpu-stress-prod"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
            <p className="text-muted-foreground text-xs">
              Lowercase alphanumeric with optional hyphens, 1–63 chars.
            </p>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="description">Description (optional)</Label>
            <Textarea
              id="description"
              placeholder="What does this scenario test?"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="fault-type">Fault type</Label>
            <Select
              value={faultType}
              onValueChange={(v) => setFaultType(v as FaultType)}
            >
              <SelectTrigger id="fault-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {FAULT_TYPES.map((t) => (
                  <SelectItem key={t} value={t}>
                    {t}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="duration">Duration (seconds)</Label>
            <Input
              id="duration"
              type="number"
              min={1}
              max={3600}
              value={duration}
              onChange={(e) => setDuration(Number(e.target.value))}
            />
          </div>

          {error ? (
            <div className="border-destructive bg-destructive/10 text-destructive rounded-md border p-3 text-sm">
              {error}
            </div>
          ) : null}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => setOpen(false)}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button onClick={() => void handleSubmit()} disabled={isSubmitting}>
            {isSubmitting ? "Creating…" : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
