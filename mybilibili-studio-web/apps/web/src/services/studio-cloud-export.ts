import { mybilibiliFetch } from "./mybilibili-api";

export interface StudioExportTask {
  taskId: string;
  projectId?: string;
  projectName?: string;
  status: "PENDING" | "RUNNING" | "SUCCEEDED" | "FAILED" | "CANCELLED";
  progress: number;
  stage?: string;
  message?: string;
  errorMessage?: string;
  outputUrl?: string;
  createdAt: number;
  updatedAt: number;
}

export interface CloudExportSettings {
  format: "mp4";
  codec: "h264";
  width: number;
  height: number;
  frameRate: number;
  bitrate: number;
  quality: number;
}

function stripLocalOnlyProjectData(project: any): any {
  const localOnlyKeys = new Set([
    "blob",
    "fileHandle",
    "waveformData",
    "filmstripThumbnails",
  ]);

  return JSON.parse(
    JSON.stringify(project, (key, value) => {
      if (localOnlyKeys.has(key)) return null;
      if (typeof value === "string" && value.startsWith("blob:")) return null;
      return value;
    }),
  );
}

export async function createStudioCloudExportTask(
  project: any,
  exportSettings: CloudExportSettings,
): Promise<StudioExportTask> {
  const cleanedProject = stripLocalOnlyProjectData(project);
  const result = await mybilibiliFetch<StudioExportTask>("/studio/export-tasks", {
    method: "POST",
    body: JSON.stringify({
      projectId: project.id,
      projectName: project.name,
      project: cleanedProject,
      exportSettings,
    }),
  });

  if (result.code !== 200 || !result.data?.taskId) {
    throw new Error(result.message || "云端导出任务创建失败");
  }
  return result.data;
}

export async function getStudioCloudExportTask(taskId: string): Promise<StudioExportTask> {
  const result = await mybilibiliFetch<StudioExportTask>(`/studio/export-tasks/${taskId}`);
  if (result.code !== 200 || !result.data?.taskId) {
    throw new Error(result.message || "云端导出任务获取失败");
  }
  return result.data;
}

export async function cancelStudioCloudExportTask(taskId: string): Promise<StudioExportTask> {
  const result = await mybilibiliFetch<StudioExportTask>(`/studio/export-tasks/${taskId}/cancel`, {
    method: "POST",
  });
  if (result.code !== 200 || !result.data?.taskId) {
    throw new Error(result.message || "云端导出任务取消失败");
  }
  return result.data;
}
