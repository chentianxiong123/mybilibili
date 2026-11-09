import { mybilibiliFetch } from "./mybilibili-api";

export interface StudioCloudAsset {
  mediaId: string;
  objectKey: string;
  url: string;
  fileName: string;
  contentType: string;
  size: number;
}

export async function uploadStudioCloudAsset(
  projectId: string,
  mediaId: string,
  file: File,
): Promise<StudioCloudAsset> {
  const form = new FormData();
  form.append("projectId", projectId);
  form.append("mediaId", mediaId);
  form.append("file", file, file.name);

  const result = await mybilibiliFetch<StudioCloudAsset>("/studio/assets/upload", {
    method: "POST",
    body: form,
  });

  if (result.code !== 200 || !result.data?.objectKey) {
    throw new Error(result.message || "素材云端上传失败");
  }
  return result.data;
}
