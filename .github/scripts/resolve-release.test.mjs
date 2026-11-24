import test from "node:test";
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { resolveReleasePlan } from "./resolve-release.mjs";
import { detectCiChanges } from "./detect-ci-changes.mjs";

function createRepo() {
  const dir = mkdtempSync(join(tmpdir(), "airvoice-release-test-"));
  const run = (...args) =>
    execFileSync("git", args, {
      cwd: dir,
      encoding: "utf8",
      env: {
        ...process.env,
        GIT_AUTHOR_NAME: "Test",
        GIT_AUTHOR_EMAIL: "test@example.com",
        GIT_COMMITTER_NAME: "Test",
        GIT_COMMITTER_EMAIL: "test@example.com",
      },
    }).trim();

  run("init", "-b", "main");

  const commitFile = (relPath, content, message) => {
    const full = join(dir, relPath);
    mkdirSync(dirname(full), { recursive: true });
    writeFileSync(full, content);
    run("add", relPath);
    run("commit", "-m", message);
    return run("rev-parse", "HEAD");
  };

  return {
    dir,
    run,
    commitFile,
    cleanup: () => rmSync(dir, { recursive: true, force: true }),
  };
}

test("auto release bumps patch on app change and skips iOS when ios/ untouched", () => {
  const repo = createRepo();
  try {
    repo.commitFile("cli/main.go", "package main\n", "init");
    repo.run("tag", "v0.4.5");
    repo.commitFile("cli/main.go", "package main // fix\n", "🐛 [cli]: Fix bug");

    const plan = resolveReleasePlan({ cwd: repo.dir, eventName: "workflow_run" });
    assert.equal(plan.skip, false);
    assert.equal(plan.version, "0.4.6");
    assert.equal(plan.tag, "v0.4.6");
    assert.equal(plan.createTag, true);
    assert.equal(plan.iosChanged, false);
    assert.equal(plan.bump, "patch");
  } finally {
    repo.cleanup();
  }
});

test("auto release sets iosChanged=true when ios/ changes", () => {
  const repo = createRepo();
  try {
    repo.commitFile("ios/App.swift", "// v1\n", "init");
    repo.run("tag", "v0.4.5");
    repo.commitFile("ios/App.swift", "// v2\n", "✨ [ios]: Improve mic button");

    const plan = resolveReleasePlan({ cwd: repo.dir, eventName: "workflow_run" });
    assert.equal(plan.skip, false);
    assert.equal(plan.version, "0.4.6");
    assert.equal(plan.iosChanged, true);
  } finally {
    repo.cleanup();
  }
});

test("auto release skips docs, www, and Formula-only commits", () => {
  const repo = createRepo();
  try {
    repo.commitFile("cli/main.go", "package main\n", "init");
    repo.run("tag", "v0.4.5");
    repo.commitFile("README.md", "# Airvoice\n", "📝 Update README");
    repo.commitFile("www/src/index.astro", "<div/>\n", "🎨 Update www");
    repo.commitFile("Formula/airvoice.rb", "# formula\n", "🔖 [homebrew]: Update formula");

    const plan = resolveReleasePlan({ cwd: repo.dir, eventName: "workflow_run" });
    assert.equal(plan.skip, true);
  } finally {
    repo.cleanup();
  }
});

test("auto release honors [major] > [minor] > patch across range and ignores [skip release]", () => {
  const repo = createRepo();
  try {
    repo.commitFile("cli/main.go", "v1\n", "init");
    repo.run("tag", "v0.4.5");
    repo.commitFile("cli/main.go", "v2\n", "✨ [cli]: New feature [minor]");
    repo.commitFile("desktop/app.go", "v2\n", "💥 [desktop]: Breaking change [major]");
    repo.commitFile("README.md", "docs\n", "📝 Update docs");

    const plan = resolveReleasePlan({ cwd: repo.dir, eventName: "workflow_run" });
    assert.equal(plan.skip, false);
    assert.equal(plan.version, "1.0.0");
    assert.equal(plan.bump, "major");

    // HEAD with [skip release] skips even when unreleased code exists
    repo.commitFile("cli/main.go", "v3\n", "🔧 Chore [skip release]");
    const skipped = resolveReleasePlan({ cwd: repo.dir, eventName: "workflow_run" });
    assert.equal(skipped.skip, true);
  } finally {
    repo.cleanup();
  }
});

test("auto release skips stale commit that is an ancestor of highest tag", () => {
  const repo = createRepo();
  try {
    repo.commitFile("cli/main.go", "v1\n", "init");
    repo.run("tag", "v0.4.5");
    const c1 = repo.commitFile("desktop/app.go", "v2\n", "fix desktop");
    repo.commitFile("cli/main.go", "v2\n", "fix cli");
    repo.run("tag", "v0.4.6");

    const plan = resolveReleasePlan({
      cwd: repo.dir,
      eventName: "workflow_run",
      targetSha: c1,
    });
    assert.equal(plan.skip, true);
    assert.match(plan.reason, /ancestor of v0\.4\.6/);
  } finally {
    repo.cleanup();
  }
});

test("manual workflow_dispatch supports minor, explicit version, and ios modes", () => {
  const repo = createRepo();
  try {
    repo.commitFile("cli/main.go", "v1\n", "init");
    repo.run("tag", "v0.4.5");

    const minorPlan = resolveReleasePlan({
      cwd: repo.dir,
      eventName: "workflow_dispatch",
      inputBump: "minor",
      inputIos: "always",
    });
    assert.equal(minorPlan.skip, false);
    assert.equal(minorPlan.version, "0.5.0");
    assert.equal(minorPlan.iosChanged, true);

    const explicitPlan = resolveReleasePlan({
      cwd: repo.dir,
      eventName: "workflow_dispatch",
      inputVersion: "2.0.0",
      inputIos: "skip",
    });
    assert.equal(explicitPlan.skip, false);
    assert.equal(explicitPlan.version, "2.0.0");
    assert.equal(explicitPlan.iosChanged, false);
  } finally {
    repo.cleanup();
  }
});

test("manual tag push reuses existing tag without re-creating it", () => {
  const repo = createRepo();
  try {
    repo.commitFile("cli/main.go", "v1\n", "init");
    repo.run("tag", "v0.5.0");

    const plan = resolveReleasePlan({
      cwd: repo.dir,
      eventName: "push",
      refName: "v0.5.0",
    });
    assert.equal(plan.skip, false);
    assert.equal(plan.version, "0.5.0");
    assert.equal(plan.tag, "v0.5.0");
    assert.equal(plan.createTag, false);
    assert.equal(plan.iosChanged, true);
  } finally {
    repo.cleanup();
  }
});

test("detectCiChanges checks all unreleased commits since latest tag on push to main", () => {
  const repo = createRepo();
  try {
    repo.commitFile("cli/main.go", "v1\n", "init");
    repo.run("tag", "v0.4.5");
    repo.commitFile("desktop/app.go", "v2\n", "desktop change");
    repo.commitFile("cli/main.go", "v2\n", "cli change");

    const scopes = detectCiChanges({ cwd: repo.dir, eventName: "push" });
    assert.equal(scopes.cli, true);
    assert.equal(scopes.desktop, true);
    assert.equal(scopes.android, false);
    assert.equal(scopes.ios, false);
    assert.equal(scopes.www, false);
  } finally {
    repo.cleanup();
  }
});

test("detectCiChanges enables ios scope when ios/ files change", () => {
  const repo = createRepo();
  try {
    repo.commitFile("ios/Airvoice/Services/ConnectionManager.swift", "// v1\n", "init");
    repo.run("tag", "v0.4.5");
    repo.commitFile("ios/Airvoice/Services/ConnectionManager.swift", "// v2\n", "update ios");

    const scopes = detectCiChanges({ cwd: repo.dir, eventName: "push" });
    assert.equal(scopes.cli, false);
    assert.equal(scopes.desktop, false);
    assert.equal(scopes.android, false);
    assert.equal(scopes.ios, true);
    assert.equal(scopes.www, false);
  } finally {
    repo.cleanup();
  }
});
