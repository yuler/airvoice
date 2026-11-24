import { appendFileSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { pathToFileURL } from "node:url";

const STRICT_SEMVER_RE = /^v?(\d+)\.(\d+)\.(\d+)$/;
const FLEX_SEMVER_RE = /^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?$/;
const BUMP_RANK = { patch: 0, minor: 1, major: 2 };

export function parseSemver(version) {
  const m = STRICT_SEMVER_RE.exec(String(version || "").trim());
  return m ? [Number(m[1]), Number(m[2]), Number(m[3])] : null;
}

export function isValidVersion(version) {
  return FLEX_SEMVER_RE.test(String(version || "").trim());
}

export function normalizeVersion(version) {
  return String(version || "").trim().replace(/^v/, "");
}

export function cmpSemver(a, b) {
  for (let i = 0; i < 3; i++) {
    if (a[i] !== b[i]) return a[i] - b[i];
  }
  return 0;
}

export function fmtSemver([maj, min, pat]) {
  return `${maj}.${min}.${pat}`;
}

export function bumpSemver([maj, min, pat], kind) {
  if (kind === "major") return [maj + 1, 0, 0];
  if (kind === "minor") return [maj, min + 1, 0];
  return [maj, min, pat + 1];
}

export function isIosPath(file) {
  return (
    file.startsWith("ios/") ||
    file === "scripts/ios-xcodegen.sh" ||
    file === "scripts/lib.sh" ||
    file === "scripts/resolve-version.sh" ||
    file === ".github/workflows/ios-release.yml"
  );
}

export function isReleasePath(file) {
  return (
    file.startsWith("cli/") ||
    file.startsWith("desktop/") ||
    file.startsWith("android/") ||
    file.startsWith("ios/") ||
    file === "go.mod" ||
    file === "go.sum" ||
    file === "scripts/lib.sh" ||
    file === "scripts/resolve-version.sh" ||
    file === "scripts/with-wails-version.sh" ||
    file === "scripts/ios-xcodegen.sh" ||
    file === ".github/workflows/cli-release.yml" ||
    file === ".github/workflows/desktop-release.yml" ||
    file === ".github/workflows/android-release.yml" ||
    file === ".github/workflows/ios-release.yml"
  );
}

export function hasSkipRelease(subject) {
  return /\[skip release\]/i.test(String(subject || ""));
}

export function detectBumpFromMessage(message) {
  const text = String(message || "");
  if (/\[major\]/i.test(text)) return "major";
  if (/\[minor\]/i.test(text)) return "minor";
  return "patch";
}

export function pickHigherBump(a, b) {
  return (BUMP_RANK[b] ?? 0) > (BUMP_RANK[a] ?? 0) ? b : a;
}

function git(args, cwd) {
  return execFileSync("git", args, {
    cwd,
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  }).trim();
}

function gitOk(args, cwd) {
  try {
    execFileSync("git", args, { cwd, stdio: "ignore" });
    return true;
  } catch {
    return false;
  }
}

export function listSemverTags(cwd, { mergedRef = null } = {}) {
  const args = ["tag", "--list", "v*"];
  if (mergedRef) {
    args.push("--merged", mergedRef);
  }
  const out = git(args, cwd);
  if (!out) return [];
  const tags = [];
  for (const line of out.split(/\r?\n/)) {
    const tag = line.trim();
    const semver = parseSemver(tag);
    if (semver && tag === `v${fmtSemver(semver)}`) {
      tags.push({ tag, semver });
    }
  }
  tags.sort((a, b) => cmpSemver(b.semver, a.semver));
  return tags;
}

function getCommitChangedFiles(hash, cwd) {
  const hasParent = gitOk(["rev-parse", "--verify", `${hash}^1`], cwd);
  const out = hasParent
    ? git(["diff", "--name-only", `${hash}^1`, hash], cwd)
    : git(["diff-tree", "--no-commit-id", "--name-only", "-r", "--root", hash], cwd);
  return out ? out.split(/\r?\n/).map((s) => s.trim()).filter(Boolean) : [];
}

function parseLogRecords(raw) {
  if (!raw) return [];
  return raw
    .split("\x1e")
    .map((entry) => entry.replace(/^\n+/, ""))
    .filter(Boolean)
    .map((entry) => {
      const [hash = "", subject = "", body = ""] = entry.split("\x1f");
      return { hash: hash.trim(), subject: subject.trim(), body };
    });
}

export function collectUnreleasedChanges(cwd, baseTag, targetSha) {
  const range = baseTag ? `${baseTag}..${targetSha}` : targetSha;

  // First-parent walk for changed files (captures full PR diffs cleanly while
  // allowing direct [skip release] commits on main to be excluded).
  const firstParentRaw = git(
    ["log", "--first-parent", "--format=%H%x1f%s%x1f%B%x1e", range],
    cwd
  );
  const firstParentCommits = parseLogRecords(firstParentRaw);

  const changedFiles = new Set();
  for (const commit of firstParentCommits) {
    if (hasSkipRelease(commit.subject)) continue;
    for (const file of getCommitChangedFiles(commit.hash, cwd)) {
      changedFiles.add(file);
    }
  }

  // Full commit walk for bump markers so [minor]/[major] works in Squash,
  // Rebase, and Merge Commit workflows alike.
  const allRaw = git(["log", "--format=%H%x1f%s%x1f%B%x1e", range], cwd);
  const allCommits = parseLogRecords(allRaw);

  let bump = "patch";
  for (const commit of allCommits) {
    if (hasSkipRelease(commit.subject)) continue;
    bump = pickHigherBump(bump, detectBumpFromMessage(commit.body));
  }

  return {
    commitCount: firstParentCommits.length,
    changedFiles: [...changedFiles],
    bump,
  };
}

export function resolveReleasePlan({
  cwd = process.cwd(),
  eventName = process.env.GITHUB_EVENT_NAME || "workflow_run",
  refName = process.env.GITHUB_REF_NAME || "",
  targetSha = process.env.TARGET_SHA || "",
  inputBump = process.env.INPUT_BUMP || "patch",
  inputVersion = process.env.INPUT_VERSION || "",
  inputIos = process.env.INPUT_IOS || "always",
} = {}) {
  const sha = targetSha ? git(["rev-parse", targetSha], cwd) : git(["rev-parse", "HEAD"], cwd);
  const allTags = listSemverTags(cwd);
  const highestTag = allTags[0] || null;
  const mergedTags = listSemverTags(cwd, { mergedRef: sha });
  const baseTag = mergedTags[0] || null;

  // 1. Manual git push of a v* tag
  if (eventName === "push") {
    const cleaned = normalizeVersion(refName);
    if (!isValidVersion(cleaned)) {
      throw new Error(`Invalid release tag: ${refName} (expected vX.Y.Z)`);
    }
    return {
      skip: false,
      version: cleaned,
      tag: `v${cleaned}`,
      createTag: false,
      iosChanged: true,
      sha,
      bump: "manual-tag",
      reason: `manual tag push v${cleaned}`,
    };
  }

  // 2. Manual workflow_dispatch
  if (eventName === "workflow_dispatch") {
    let version;
    let bump = inputBump || "patch";
    if (!["patch", "minor", "major"].includes(bump)) {
      throw new Error(`Invalid bump input: ${bump}`);
    }

    if (inputVersion && inputVersion.trim()) {
      version = normalizeVersion(inputVersion);
      if (!isValidVersion(version)) {
        throw new Error(`Invalid explicit version: ${inputVersion}`);
      }
      if (gitOk(["rev-parse", "--verify", `refs/tags/v${version}`], cwd)) {
        throw new Error(`Tag v${version} already exists`);
      }
      bump = "explicit";
    } else {
      const base = highestTag ? highestTag.semver : [0, 0, 0];
      version = fmtSemver(bumpSemver(base, bump));
    }

    let iosChanged = true;
    if (inputIos === "skip") {
      iosChanged = false;
    } else if (inputIos === "auto") {
      const changes = collectUnreleasedChanges(cwd, baseTag?.tag || null, sha);
      iosChanged = changes.changedFiles.some(isIosPath);
    }

    return {
      skip: false,
      version,
      tag: `v${version}`,
      createTag: true,
      iosChanged,
      sha,
      bump,
      reason: `manual workflow_dispatch (${bump})`,
    };
  }

  // 3. Automatic release after CI succeeds on push to main (workflow_run)
  const headSubject = git(["log", "-1", "--format=%s", sha], cwd);
  if (hasSkipRelease(headSubject)) {
    return {
      skip: true,
      version: "",
      tag: "",
      createTag: false,
      iosChanged: false,
      sha,
      bump: "patch",
      reason: "HEAD commit subject contains [skip release]",
    };
  }

  const headTags = git(["tag", "--points-at", sha], cwd)
    .split(/\r?\n/)
    .map((t) => t.trim())
    .filter((t) => parseSemver(t) !== null);
  if (headTags.length > 0) {
    return {
      skip: true,
      version: normalizeVersion(headTags[0]),
      tag: headTags[0],
      createTag: false,
      iosChanged: false,
      sha,
      bump: "patch",
      reason: `HEAD is already tagged as ${headTags[0]}`,
    };
  }

  // Prevent out-of-order/stale CI runs from releasing an older commit
  if (highestTag && gitOk(["merge-base", "--is-ancestor", sha, highestTag.tag], cwd)) {
    return {
      skip: true,
      version: "",
      tag: "",
      createTag: false,
      iosChanged: false,
      sha,
      bump: "patch",
      reason: `commit ${sha.slice(0, 7)} is already an ancestor of ${highestTag.tag}`,
    };
  }

  const changes = collectUnreleasedChanges(cwd, baseTag?.tag || null, sha);
  if (changes.commitCount === 0) {
    return {
      skip: true,
      version: "",
      tag: "",
      createTag: false,
      iosChanged: false,
      sha,
      bump: "patch",
      reason: `no new commits since ${baseTag?.tag || "initial commit"}`,
    };
  }

  const matchedReleaseFiles = changes.changedFiles.filter(isReleasePath);
  if (matchedReleaseFiles.length === 0) {
    return {
      skip: true,
      version: "",
      tag: "",
      createTag: false,
      iosChanged: false,
      sha,
      bump: changes.bump,
      reason: `no release-affecting files changed since ${baseTag?.tag || "initial commit"}`,
    };
  }

  const iosChanged = changes.changedFiles.some(isIosPath);
  const base = highestTag ? highestTag.semver : [0, 0, 0];
  const nextVersion = fmtSemver(bumpSemver(base, changes.bump));

  return {
    skip: false,
    version: nextVersion,
    tag: `v${nextVersion}`,
    createTag: true,
    iosChanged,
    sha,
    bump: changes.bump,
    reason: `matched ${matchedReleaseFiles.slice(0, 5).join(", ")} since ${baseTag?.tag || "root"} (bump=${changes.bump}, ios=${iosChanged})`,
  };
}

function writeGitHubOutput(plan) {
  if (!process.env.GITHUB_OUTPUT) return;
  const lines = [
    `skip=${plan.skip}`,
    `version=${plan.version}`,
    `tag=${plan.tag}`,
    `create_tag=${plan.createTag}`,
    `ios_changed=${plan.iosChanged}`,
    `sha=${plan.sha}`,
    `bump=${plan.bump}`,
  ];
  appendFileSync(process.env.GITHUB_OUTPUT, `${lines.join("\n")}\n`);
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  const plan = resolveReleasePlan();
  console.log(
    plan.skip
      ? `Skipping release: ${plan.reason}`
      : `Planned release ${plan.tag} at ${plan.sha.slice(0, 7)}: ${plan.reason}`
  );
  writeGitHubOutput(plan);
}
