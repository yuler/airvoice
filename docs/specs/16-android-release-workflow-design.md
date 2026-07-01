# Design Spec: Android Release Workflow

## Goal

Generate a signed Android release APK automatically from GitHub Actions while keeping signing material out of git.

## Release Signing

The Android `release` build reads signing settings from environment variables:

- `ANDROID_KEYSTORE_PATH`
- `ANDROID_KEYSTORE_PASSWORD`
- `ANDROID_KEY_ALIAS`
- `ANDROID_KEY_PASSWORD`

When all four variables are present, Gradle signs the release APK. When they are absent, local `assembleRelease` behavior remains available for development and produces the normal unsigned release output.

## GitHub Actions Workflow

Create `.github/workflows/android-release.yml` with these triggers:

- `push` to `main` when Android or workflow files change
- `push` of tags matching `v*`
- manual `workflow_dispatch`

The workflow runs on Ubuntu, sets up JDK 17, caches Gradle, validates required signing secrets, decodes the keystore into `$RUNNER_TEMP`, runs Android unit tests, builds the signed release APK, uploads the APK artifact, and attaches it to a GitHub Release for tag builds.

## GitHub Secrets

The workflow uses these repository secrets:

- `ANDROID_KEYSTORE_BASE64`
- `ANDROID_KEYSTORE_PASSWORD`
- `ANDROID_KEY_ALIAS`
- `ANDROID_KEY_PASSWORD`

`ANDROID_KEYSTORE_BASE64` is the base64-encoded local keystore file. The password and alias secrets must match the generated keystore.

## Local Signing Files

Generate the keystore in the current repository folder as `airvoice-release.keystore`. Generate a local helper file named `airvoice-release-secrets.env` containing the GitHub secret values. Both files are ignored by git.

## Verification

Verify the change by running:

```bash
cd android && ./gradlew testDebugUnitTest assembleRelease
```

Run the command with the generated environment variables exported to confirm the APK is signed locally before relying on CI.
