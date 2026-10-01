import { getPreferenceValues } from "@raycast/api";
import { execFile } from "node:child_process";
import { accessSync, constants } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";

export interface Host {
  alias: string;
  hostname: string;
  user: string;
  count: number;
}

export interface SendResult {
  host: string;
  dir: string;
  remote_paths: string[];
}

export interface Destination {
  host: string;
  dir: string;
}

// Raycast starts extensions with a minimal PATH; the binary itself needs
// ssh, tar and sh, which live in the system directories.
const PATH = ["/usr/local/bin", "/opt/homebrew/bin", "/usr/bin", "/bin", "/usr/sbin", "/sbin"].join(":");

function candidates(): string[] {
  const home = homedir();
  return [
    "/usr/local/bin/drop-box-cli",
    join(home, ".local/bin/drop-box-cli"),
    join(home, "go/bin/drop-box-cli"),
    "/opt/homebrew/bin/drop-box-cli",
  ];
}

function isExecutable(path: string): boolean {
  try {
    accessSync(path, constants.X_OK);
    return true;
  } catch {
    return false;
  }
}

function binaryPath(): string {
  const { binaryPath } = getPreferenceValues<{ binaryPath?: string }>();
  const configured = binaryPath?.trim().replace(/^~(?=\/|$)/, homedir());
  if (configured) {
    if (!isExecutable(configured)) {
      throw new Error(`${configured} is not an executable — check the extension preferences`);
    }
    return configured;
  }
  const found = candidates().find(isExecutable);
  if (!found) {
    throw new Error("drop-box-cli not found — install it or set its path in the extension preferences");
  }
  return found;
}

function run(args: string[]): Promise<string> {
  const bin = binaryPath();
  return new Promise((resolve, reject) => {
    execFile(
      bin,
      args,
      { env: { ...process.env, PATH: `${PATH}:${process.env.PATH ?? ""}` }, maxBuffer: 16 * 1024 * 1024 },
      (error, stdout, stderr) => {
        if (error) {
          const msg = stderr.trim().replace(/^drop-box-cli: /, "");
          reject(new Error(msg || error.message));
        } else {
          resolve(stdout);
        }
      },
    );
  });
}

export async function listHosts(): Promise<Host[]> {
  return JSON.parse(await run(["hosts", "--json"]));
}

export async function listDirs(host: string): Promise<string[]> {
  return JSON.parse(await run(["dirs", "--host", host, "--json"]));
}

export async function lastDestination(): Promise<Destination> {
  return JSON.parse(await run(["last", "--json"]));
}

export async function send(host: string, dir: string, files: string[]): Promise<SendResult> {
  // The extension copies to the clipboard itself, via Raycast.
  return JSON.parse(await run(["send", "--host", host, "--dir", dir, "--json", "--no-clipboard", "--", ...files]));
}
