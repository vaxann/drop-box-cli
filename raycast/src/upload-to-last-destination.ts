import { showHUD } from "@raycast/api";
import { lastDestination } from "./cli";
import { errorMessage, selectedFiles, upload } from "./upload";

export default async function Command() {
  let files: string[];
  let host: string;
  let dir: string;
  try {
    files = await selectedFiles();
    ({ host, dir } = await lastDestination());
  } catch (error) {
    const message = errorMessage(error);
    await showHUD(
      message === "nothing has been sent yet" ? "No previous upload — use Upload Finder Selection first" : message,
    );
    return;
  }
  await upload(host, dir, files);
}
