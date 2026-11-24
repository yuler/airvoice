import { appendFileSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { pathToFileURL } from "node:url";
import { collectUnreleasedChanges, listSemverTags } from "./resolve-release.mjs";

export function matchScope(file) {
  return {
    cli:
      file.startsWith("cli/") ||
      file === "go.mod" ||
      file === "go.sum" ||
      file === "scripts/resolve-version.sh" ||
      file === "scripts/lib.sh" ||
      file === ".github/workflows/ci.yml" ||
      file === ".github/scripts/detect-ci-changes.mjs" ||
      file === ".github/scripts/resolve-release.mjs",
    desktop:
      file.startsWith("desktop/") ||
      file === "go.mod" ||
      file === "go.sum" ||
      file === "scripts/resolve-version.sh" ||
      file === "scripts/lib.sh" ||
      file === "scripts/with-wails-version.sh" ||
      file === ".github/workflows/ci.yml" ||
      file === ".github/scripts/detect-ci-changes.mjs" ||
      file === ".github/scripts/resolve-release.mjs",
    android:
      file.startsWith("android/") ||
      file === "scripts/resolve-version.sh" ||
      file === ".github/workflows/ci.yml" ||
      file === ".github/scripts/detect-ci-changes.mjs" ||
      file === ".github/scripts/resolve-release.mjs",
    ios:
      file.startsWith("ios/") ||
      file === "scripts/ios-xcodegen.sh" ||
      file === "scripts/resolve-version.sh" ||
      file === "scripts/lib.sh" ||
      file === ".github/workflows/ci.yml" ||
      file === ".github/scripts/detect-ci-changes.mjs" ||
      file === ".github/scripts/resolve-release.mjs",
    www:
      file.startsWith("www/") ||
      file === "scripts/resolve-version.sh" ||
      file === ".github/workflows/ci.yml" ||
      file === ".github/scripts/detect-ci-changes.mjs" ||
      file === ".github/scripts/resolve-release.mjs",
  };
}

function git(args, cwd) {
  return execFileSync("git", args, {
    cwd,
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  }).trim();
}

export function detectCiChanges({
  cwd = process.cwd(),
  eventName = process.env.GITHUB_EVENT_NAME || "push",
  prBaseSha = process.env.PR_BASE_SHA || "",
  targetSha = process.env.TARGET_SHA || "HEAD",
} = {}) {
  if (eventName === "workflow_dispatch") {
    return {
      cli: true,
      desktop: true,
      android: true,
      ios: true,
      www: true,
      reason: "manual workflow_dispatch runs all checks",
    };
  }

  let files = [];
  let reason = "";

  if (eventName === "pull_request" && prBaseSha) {
    const out = git(["diff", "--name-only", `${prBaseSha}...${targetSha}`], cwd);
    files = out ? out.split(/\r?\n/).map((s) => s.trim()).filter(Boolean) : [];
    reason = `PR diff ${prBaseSha.slice(0, 7)}...${targetSha.slice(0, 7)} (${files.length} files)`;
  } else {
    // On push to main, inspect ALL unreleased changes since the latest
    // reachable v* tag (excluding [skip release] commits) so cancelling
    // superseded CI runs never skips testing an earlier unreleased commit.
    const sha = git(["rev-parse", targetSha], cwd);
    const mergedTags = listSemverTags(cwd, { mergedRef: sha });
    const baseTag = mergedTags[0]?.tag || null;
    const changes = collectUnreleasedChanges(cwd, baseTag, sha);
    files = changes.changedFiles;
    reason = `unreleased diff since ${baseTag || "root"} (${files.length} files)`;
  }

  const scopes = { cli: false, desktop: false, android: false, ios: false, www: false };
  for (const file of files) {
    const matched = matchScope(file);
    for (const key of Object.keys(scopes)) {
      if (matched[key]) scopes[key] = true;
    }
  }

  return { ...scopes, reason };
}

function writeGitHubOutput(scopes) {
  if (!process.env.GITHUB_OUTPUT) return;
  const lines = [
    `cli=${scopes.cli}`,
    `desktop=${scopes.desktop}`,
    `android=${scopes.android}`,
    `ios=${scopes.ios}`,
    `www=${scopes.www}`,
  ];
  appendFileSync(process.env.GITHUB_OUTPUT, `${lines.join("\n")}\n`);
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  const scopes = detectCiChanges();
  console.log(
    `Detected CI scopes (${scopes.reason}): cli=${scopes.cli} desktop=${scopes.desktop} android=${scopes.android} ios=${scopes.ios} www=${scopes.www}`
  );
  writeGitHubOutput(scopes);
}
