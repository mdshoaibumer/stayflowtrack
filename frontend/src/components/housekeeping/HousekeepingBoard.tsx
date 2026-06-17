"use client";

import React, { useState, useEffect, useCallback } from "react";
import { useApi } from "@/hooks/useApi";

interface Task {
  id: string;
  unit_number: string;
  assignee_name: string;
  status: string;
  priority: string;
  task_type: string;
  notes: string;
  estimated_minutes: number;
  created_at: string;
}

interface HousekeepingBoardProps {
  propertyId: string;
}

const statusLabels: Record<string, string> = {
  dirty: "Dirty",
  cleaning: "Cleaning",
  inspection: "Inspection",
  ready: "Ready",
};

const statusColors: Record<string, string> = {
  dirty: "bg-red-100 border-red-300",
  cleaning: "bg-yellow-100 border-yellow-300",
  inspection: "bg-blue-100 border-blue-300",
  ready: "bg-green-100 border-green-300",
};

const priorityBadge: Record<string, string> = {
  urgent: "bg-red-600 text-white",
  high: "bg-orange-500 text-white",
  normal: "bg-gray-200 text-gray-700",
  low: "bg-gray-100 text-gray-500",
};

export default function HousekeepingBoard({ propertyId }: HousekeepingBoardProps) {
  const api = useApi();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<string>("");

  const fetchTasks = useCallback(async () => {
    const params: Record<string, string> = { property_id: propertyId, per_page: "100" };
    if (filter) params.status = filter;

    try {
      const data = await api.get<any>("/api/v1/housekeeping/tasks", params);
      setTasks(Array.isArray(data) ? data : data?.data || []);
    } catch {
      // Error handled by useApi (401 logout)
    }
    setLoading(false);
  }, [api, propertyId, filter]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const updateStatus = async (taskId: string, newStatus: string) => {
    try {
      await api.post("/api/v1/housekeeping/tasks/status", { task_id: taskId, status: newStatus });
      fetchTasks();
    } catch {
      // Error handled by useApi
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center h-32"><div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div></div>;
  }

  // Group by status for kanban view
  const columns = ["dirty", "cleaning", "inspection", "ready"];
  const grouped = columns.reduce((acc, status) => {
    acc[status] = tasks.filter((t) => t.status === status);
    return acc;
  }, {} as Record<string, Task[]>);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">Housekeeping</h2>
        <div className="flex gap-2">
          <button onClick={() => setFilter("")} className={`px-3 py-1 text-xs rounded ${!filter ? "bg-blue-600 text-white" : "border"}`}>All</button>
          {columns.map((s) => (
            <button key={s} onClick={() => setFilter(s)} className={`px-3 py-1 text-xs rounded capitalize ${filter === s ? "bg-blue-600 text-white" : "border"}`}>
              {s}
            </button>
          ))}
        </div>
      </div>

      {/* Kanban Board */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {columns.map((status) => (
          <div key={status} className={`rounded-lg border-2 p-3 ${statusColors[status]}`}>
            <div className="flex items-center justify-between mb-3">
              <h3 className="font-semibold text-sm">{statusLabels[status]}</h3>
              <span className="text-xs bg-white px-2 py-0.5 rounded-full font-medium">{grouped[status]?.length || 0}</span>
            </div>
            <div className="space-y-2 max-h-[60vh] overflow-y-auto">
              {(grouped[status] || []).map((task) => (
                <TaskCard key={task.id} task={task} onStatusChange={updateStatus} currentStatus={status} />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function TaskCard({ task, onStatusChange, currentStatus }: { task: Task; onStatusChange: (id: string, status: string) => void; currentStatus: string }) {
  const nextStatus: Record<string, string> = { dirty: "cleaning", cleaning: "inspection", inspection: "ready" };
  const next = nextStatus[currentStatus];

  return (
    <div className="bg-white rounded p-3 shadow-sm border text-sm">
      <div className="flex items-center justify-between mb-1">
        <span className="font-bold">{task.unit_number}</span>
        <span className={`text-[10px] px-1.5 py-0.5 rounded ${priorityBadge[task.priority]}`}>{task.priority}</span>
      </div>
      <div className="text-xs text-gray-500 mb-1">{task.task_type.replaceAll("_", " ")}</div>
      {task.assignee_name && <div className="text-xs text-gray-600">👤 {task.assignee_name}</div>}
      {task.notes && <div className="text-xs text-gray-400 mt-1 truncate">{task.notes}</div>}
      {next && (
        <button
          onClick={() => onStatusChange(task.id, next)}
          className="mt-2 w-full text-xs bg-blue-600 text-white py-1 rounded hover:bg-blue-700"
        >
          Move to {statusLabels[next]}
        </button>
      )}
    </div>
  );
}
