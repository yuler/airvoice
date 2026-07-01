# Android Release Workflow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a GitHub Actions workflow that builds and uploads a signed Android release APK.

**Architecture:** Gradle owns signing behavior through environment variables. GitHub Actions reconstructs the keystore from repository secrets in runner temp storage, then runs tests and `assembleRelease`.

**Tech Stack:** Android Gradle Plugin 8.2.2, Kotlin Gradle DSL, Gradle Wrapper 8.5, GitHub Actions, JDK 17.

## Global Constraints

- Signing material must not be committed to git.
- The local keystore file is `airvoice-release.keystore`.
- The local helper secrets file is `airvoice-release-secrets.env`.
- The GitHub secret names are `ANDROID_KEYSTORE_BASE64`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`, and `ANDROID_KEY_PASSWORD`.
- Existing local unsigned `assembleRelease` behavior must remain available when signing environment variables are absent.
- Existing user changes in `mise.toml` must not be reverted.

---

### Task 1: Add Secret-Backed Release Signing

**Files:**
- Modify: `android/app/build.gradle.kts`
- Modify: `.gitignore`

**Interfaces:**
- Consumes: `ANDROID_KEYSTORE_PATH`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`, `ANDROID_KEY_PASSWORD`
- Produces: Gradle release signing config named `release` when all signing environment variables are present

- [ ] **Step 1: Update `.gitignore` for signing material**

Add:

```gitignore
# Android signing
*.keystore
*.jks
*.p12
airvoice-release-secrets.env
```

- [ ] **Step 2: Add environment-backed signing to Gradle**

Add these top-level values above the `android` block in `android/app/build.gradle.kts`:

```kotlin
val releaseKeystorePath = System.getenv("ANDROID_KEYSTORE_PATH")
val releaseKeystorePassword = System.getenv("ANDROID_KEYSTORE_PASSWORD")
val releaseKeyAlias = System.getenv("ANDROID_KEY_ALIAS")
val releaseKeyPassword = System.getenv("ANDROID_KEY_PASSWORD")
val hasReleaseSigning = listOf(
    releaseKeystorePath,
    releaseKeystorePassword,
    releaseKeyAlias,
    releaseKeyPassword
).all { !it.isNullOrBlank() }
```

Inside `android`, add a conditional `signingConfigs` block and assign it to the `release` build type when available:

```kotlin
signingConfigs {
    if (hasReleaseSigning) {
        create("release") {
            storeFile = file(requireNotNull(releaseKeystorePath))
            storePassword = requireNotNull(releaseKeystorePassword)
            keyAlias = requireNotNull(releaseKeyAlias)
            keyPassword = requireNotNull(releaseKeyPassword)
        }
    }
}
```

```kotlin
if (hasReleaseSigning) {
    signingConfig = signingConfigs.getByName("release")
}
```

- [ ] **Step 3: Verify Gradle configuration still compiles**

Run:

```bash
cd android && ./gradlew tasks --quiet
```

Expected: command exits 0 and lists Gradle tasks.

### Task 2: Add Android Release Workflow

**Files:**
- Create: `.github/workflows/android-release.yml`

**Interfaces:**
- Consumes: GitHub secrets `ANDROID_KEYSTORE_BASE64`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`, `ANDROID_KEY_PASSWORD`
- Produces: `airvoice-android-<version>.apk` artifact and GitHub Release asset on tag builds

- [ ] **Step 1: Create workflow**

Create `.github/workflows/android-release.yml`:

```yaml
name: Android Release

on:
  push:
    branches: [main]
    tags:
      - 'v*'
    paths:
      - 'android/**'
      - '.github/workflows/android-release.yml'
  workflow_dispatch:

permissions:
  contents: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up JDK 17
        uses: actions/setup-java@v4
        with:
          distribution: temurin
          java-version: '17'
          cache: gradle

      - name: Prepare release keystore
        env:
          ANDROID_KEYSTORE_BASE64: ${{ secrets.ANDROID_KEYSTORE_BASE64 }}
          ANDROID_KEYSTORE_PASSWORD: ${{ secrets.ANDROID_KEYSTORE_PASSWORD }}
          ANDROID_KEY_ALIAS: ${{ secrets.ANDROID_KEY_ALIAS }}
          ANDROID_KEY_PASSWORD: ${{ secrets.ANDROID_KEY_PASSWORD }}
        run: |
          for name in ANDROID_KEYSTORE_BASE64 ANDROID_KEYSTORE_PASSWORD ANDROID_KEY_ALIAS ANDROID_KEY_PASSWORD; do
            if [ -z "${!name}" ]; then
              echo "::error::Missing required secret: ${name}"
              exit 1
            fi
          done

          echo "$ANDROID_KEYSTORE_BASE64" | base64 --decode > "$RUNNER_TEMP/airvoice-release.keystore"

      - name: Run unit tests
        working-directory: android
        run: ./gradlew testDebugUnitTest

      - name: Build signed release APK
        working-directory: android
        env:
          ANDROID_KEYSTORE_PATH: ${{ runner.temp }}/airvoice-release.keystore
          ANDROID_KEYSTORE_PASSWORD: ${{ secrets.ANDROID_KEYSTORE_PASSWORD }}
          ANDROID_KEY_ALIAS: ${{ secrets.ANDROID_KEY_ALIAS }}
          ANDROID_KEY_PASSWORD: ${{ secrets.ANDROID_KEY_PASSWORD }}
        run: ./gradlew assembleRelease

      - name: Version
        id: version
        shell: bash
        run: |
          if [[ "${{ github.ref_type }}" == "tag" ]]; then
            echo "version=${GITHUB_REF_NAME#v}" >> "$GITHUB_OUTPUT"
          else
            echo "version=0.0.0-$(git rev-parse --short HEAD)" >> "$GITHUB_OUTPUT"
          fi

      - name: Collect APK
        shell: bash
        run: |
          mkdir -p release
          cp android/app/build/outputs/apk/release/app-release.apk "release/airvoice-android-${{ steps.version.outputs.version }}.apk"
          ls -lh release

      - name: Upload APK artifact
        uses: actions/upload-artifact@v4
        with:
          name: android-release-apk
          path: release/*.apk
          if-no-files-found: error

      - name: Attach APK to GitHub Release
        if: github.ref_type == 'tag'
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ github.ref_name }}
          name: ${{ github.ref_name }}
          draft: false
          prerelease: false
          generate_release_notes: true
          files: release/*.apk
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 2: Verify workflow syntax shape**

Run:

```bash
ruby -e "require 'yaml'; YAML.load_file('.github/workflows/android-release.yml'); puts 'ok'"
```

Expected: `ok`

### Task 3: Generate Local Release Keystore

**Files:**
- Create: `airvoice-release.keystore`
- Create: `airvoice-release-secrets.env`

**Interfaces:**
- Produces: Local keystore and exact GitHub secret values

- [ ] **Step 1: Generate passwords and keystore**

Run from the repository root:

```bash
source ./scripts/java-env.sh
STORE_PASSWORD="$(openssl rand -base64 32)"
KEY_PASSWORD="$STORE_PASSWORD"
KEY_ALIAS="airvoice-release"
"$JAVA_HOME/bin/keytool" -genkeypair \
  -v \
  -keystore airvoice-release.keystore \
  -storetype PKCS12 \
  -alias "$KEY_ALIAS" \
  -keyalg RSA \
  -keysize 4096 \
  -validity 10000 \
  -storepass "$STORE_PASSWORD" \
  -keypass "$KEY_PASSWORD" \
  -dname "CN=Airvoice, OU=Release, O=Airvoice, L=Unknown, ST=Unknown, C=US"
```

- [ ] **Step 2: Save local secret values**

Write `airvoice-release-secrets.env` from the values generated in Step 1:

```bash
umask 077
{
  printf 'ANDROID_KEYSTORE_BASE64=%s\n' "$(base64 -w 0 airvoice-release.keystore)"
  printf 'ANDROID_KEYSTORE_PASSWORD=%s\n' "$STORE_PASSWORD"
  printf 'ANDROID_KEY_ALIAS=%s\n' "$KEY_ALIAS"
  printf 'ANDROID_KEY_PASSWORD=%s\n' "$KEY_PASSWORD"
} > airvoice-release-secrets.env
```

- [ ] **Step 3: Verify signed release build locally**

Run:

```bash
set -a
source ./airvoice-release-secrets.env
set +a
export ANDROID_KEYSTORE_PATH="$PWD/airvoice-release.keystore"
cd android && ./gradlew testDebugUnitTest assembleRelease
```

Expected: command exits 0 and creates `android/app/build/outputs/apk/release/app-release.apk`.
