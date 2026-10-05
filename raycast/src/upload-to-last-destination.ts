import { LaunchProps, showHUD } from "@raycast/api";
import { lastDestination } from "./cli";
import { errorMessage, FilesContext, selectedFiles, upload } from "./upload";

export default async function Command(props: LaunchProps<{ launchContext?: FilesContext }>) {
  let files: string[];
  let host: string;
  let dir: string;
  try {
    files = await selectedFiles(props.launchContext);
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
