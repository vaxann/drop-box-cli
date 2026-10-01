import { Clipboard, getSelectedFinderItems, showHUD, showToast, Toast } from "@raycast/api";
import { basename } from "node:path";
import { send } from "./cli";

// selectedFiles returns the paths selected in the frontmost Finder window.
export async function selectedFiles(): Promise<string[]> {
  let paths: string[];
  try {
    paths = (await getSelectedFinderItems()).map((item) => item.path);
  } catch {
    throw new Error("Select files in Finder first");
  }
  if (paths.length === 0) {
    throw new Error("Select files in Finder first");
  }
  return paths;
}

export function describeFiles(files: string[]): string {
  return files.length === 1 ? `"${basename(files[0].replace(/\/+$/, ""))}"` : `${files.length} files`;
}

// upload sends files, copies the remote paths and closes Raycast with a
// HUD. It returns false (leaving a failure toast) when the upload failed.
export async function upload(host: string, dir: string, files: string[]): Promise<boolean> {
  const toast = await showToast({
    style: Toast.Style.Animated,
    title: `Uploading ${describeFiles(files)}`,
    message: `${host}:${dir}`,
  });
  try {
    const result = await send(host, dir, files);
    await Clipboard.copy(result.remote_paths.join("\n"));
    const more = result.remote_paths.length > 1 ? ` (+${result.remote_paths.length - 1} more)` : "";
    await toast.hide();
    await showHUD(`✓ ${host}:${result.remote_paths[0]}${more} — copied`);
    return true;
  } catch (error) {
    showFailure(toast, "Upload failed", error);
    return false;
  }
}

export function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export function showFailure(toast: Toast, title: string, error: unknown) {
  const message = errorMessage(error);
  toast.style = Toast.Style.Failure;
  toast.title = title;
  toast.message = message;
  toast.primaryAction = {
    title: "Copy Error",
    onAction: () => Clipboard.copy(message),
  };
}
