# Android Package Rename Implementation Plan

**Goal:** Rename the Android app package name, namespace, and application ID from `com.yule.airvoice` to `cc.yuler.airvoice` to align with the iOS Bundle ID.

**Architecture:** We will modify `build.gradle.kts` to update the Android namespace and applicationId, rename the source code directories on disk to match the new package name, and perform a global replacement in Kotlin source/test files to update package declarations and imports.

**Tech Stack:** Gradle, Kotlin, Bash (find, sed, mv).

---

## Proposed Changes

### Task 1: Update Build Configuration

**Files:**
- Modify: [build.gradle.kts](file:///home/yule/Sides/airvoice/android/app/build.gradle.kts)

- [ ] **Step 1: Update namespace and applicationId in build.gradle.kts**
  Modify namespace and applicationId to `cc.yuler.airvoice`.
  
  ```kotlin
  // build.gradle.kts
  android {
      namespace = "cc.yuler.airvoice"
      // ...
      defaultConfig {
          applicationId = "cc.yuler.airvoice"
          // ...
      }
  }
  ```

- [ ] **Step 2: Commit build configuration change**
  ```bash
  git add android/app/build.gradle.kts
  git commit -m "build: update namespace and applicationId to cc.yuler.airvoice"
  ```

---

### Task 2: Move Source Files on Disk

**Files:**
- Move: `android/app/src/main/java/com/yule/airvoice` -> `android/app/src/main/java/cc/yuler/airvoice`
- Move: `android/app/src/test/java/com/yule/airvoice` -> `android/app/src/test/java/cc/yuler/airvoice`

- [ ] **Step 1: Move main source files**
  Execute mv commands:
  ```bash
  mkdir -p android/app/src/main/java/cc/yuler
  mv android/app/src/main/java/com/yule/airvoice android/app/src/main/java/cc/yuler/
  rm -rf android/app/src/main/java/com
  ```

- [ ] **Step 2: Move test source files**
  Execute mv commands:
  ```bash
  mkdir -p android/app/src/test/java/cc/yuler
  mv android/app/src/test/java/com/yule/airvoice android/app/src/test/java/cc/yuler/
  rm -rf android/app/src/test/java/com
  ```

- [ ] **Step 3: Commit moved directories**
  ```bash
  git add android/app/src/main/java/cc android/app/src/test/java/cc
  # Note: we use git add -A or similar to stage deletions of com directory
  git add -A android/app/src/main/java/com android/app/src/test/java/com
  git commit -m "refactor: move source directories to cc/yuler/airvoice"
  ```

---

### Task 3: Replace Package Declarations and Imports

**Files:**
- Modify: All `.kt` files inside the `android/app/src/` folder.

- [ ] **Step 1: Replace com.yule.airvoice with cc.yuler.airvoice in all Kotlin files**
  Use `find` and `sed` to replace package names and imports:
  ```bash
  find android/ -name "*.kt" -type f -exec sed -i 's/com\.yule\.airvoice/cc.yuler.airvoice/g' {} +
  ```

- [ ] **Step 2: Verify changes with git diff**
  Review changes to make sure imports and packages are updated correctly.
  
- [ ] **Step 3: Commit code modifications**
  ```bash
  git add -u
  git commit -m "refactor: update package declarations and imports to cc.yuler.airvoice"
  ```

---

### Task 4: Build and Verify

- [ ] **Step 1: Run Android Unit Tests**
  Run: `mise run android:test`
  Expected: BUILD SUCCESSFUL (All tests pass).

- [ ] **Step 2: Run Android Debug Build**
  Run: `mise run android:build`
  Expected: BUILD SUCCESSFUL (App compiles successfully with the new package name).

- [ ] **Step 3: Commit final build state if any auto-generated changes occurred**
  ```bash
  git add -u
  git commit -m "build: verify compilation with new package ID"
  ```
