import { Action, ActionPanel, Icon, LaunchProps, List, showToast, Toast } from "@raycast/api";
import { usePromise } from "@raycast/utils";
import { useState } from "react";
import { Host, listDirs, listHosts } from "./cli";
import { describeFiles, errorMessage, FilesContext, selectedFiles, upload } from "./upload";

export default function Command(props: LaunchProps<{ launchContext?: FilesContext }>) {
  // Errors are rendered as empty views instead of the default toasts.
  const files = usePromise(selectedFiles, [props.launchContext], { onError: () => undefined });
  const hosts = usePromise(listHosts, [], { onError: () => undefined });

  if (files.error) {
    return (
      <List>
        <List.EmptyView icon={Icon.Finder} title={files.error.message} />
      </List>
    );
  }
  if (hosts.error) {
    return (
      <List>
        <List.EmptyView icon={Icon.ExclamationMark} title="Cannot list servers" description={hosts.error.message} />
      </List>
    );
  }

  const label = files.data ? describeFiles(files.data) : "files";
  return (
    <List isLoading={files.isLoading || hosts.isLoading} searchBarPlaceholder={`Send ${label} to which server?`}>
      <List.EmptyView
        icon={Icon.Network}
        title="No servers"
        description="Add your servers as Host entries to ~/.ssh/config"
      />
      {files.data &&
        hosts.data?.map((host) => (
          <List.Item
            key={host.alias}
            icon={Icon.HardDrive}
            title={host.alias}
            subtitle={describeHost(host)}
            accessories={host.count > 0 ? [{ text: `${host.count}×`, tooltip: "Uploads so far" }] : []}
            actions={
              <ActionPanel>
                <Action.Push
                  title="Choose Directory"
                  icon={Icon.Folder}
                  target={<DirectoryList host={host.alias} files={files.data!} />}
                />
              </ActionPanel>
            }
          />
        ))}
    </List>
  );
}

function describeHost(host: Host): string {
  return host.user && host.hostname ? `${host.user}@${host.hostname}` : host.hostname;
}

function DirectoryList({ host, files }: { host: string; files: string[] }) {
  const [search, setSearch] = useState("");
  const [uploading, setUploading] = useState(false);
  const dirs = usePromise(listDirs, [host], {
    onError: (error) => {
      showToast(Toast.Style.Failure, "Cannot list directories", errorMessage(error));
    },
  });

  async function uploadTo(dir: string) {
    if (uploading) {
      return;
    }
    setUploading(true);
    if (!(await upload(host, dir, files))) {
      setUploading(false);
    }
  }

  const typed = search.trim();
  const showTyped = typed !== "" && !dirs.data?.includes(typed);

  return (
    <List
      isLoading={dirs.isLoading || uploading}
      searchText={search}
      onSearchTextChange={setSearch}
      filtering={true}
      navigationTitle={`Send ${describeFiles(files)} to ${host}`}
      searchBarPlaceholder={`Directory on ${host} — pick or type a new path`}
    >
      <List.EmptyView icon={Icon.Folder} title="Type a directory" description="e.g. ~/inbox — created if missing" />
      {dirs.data?.map((dir) => (
        <List.Item
          key={dir}
          icon={Icon.Folder}
          title={dir}
          actions={
            <ActionPanel>
              <Action title="Upload Here" icon={Icon.Upload} onAction={() => uploadTo(dir)} />
            </ActionPanel>
          }
        />
      ))}
      {showTyped && (
        <List.Item
          key="__typed__"
          icon={Icon.NewFolder}
          title={typed}
          subtitle="new path"
          keywords={[typed]}
          actions={
            <ActionPanel>
              <Action title="Upload Here" icon={Icon.Upload} onAction={() => uploadTo(typed)} />
            </ActionPanel>
          }
        />
      )}
    </List>
  );
}
