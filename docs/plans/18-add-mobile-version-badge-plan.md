# Add Mobile Version Badge Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a version badge (e.g. `v0.3.1`) to the bottom of the home screen of both iOS and Android applications, centered and located below the input method tips card.

**Architecture:** Retrieve the application version dynamically from the application packaging configurations (via main Bundle on iOS, and PackageManager on Android) and render it centered at the bottom of the Home screen control sections.

**Tech Stack:** Swift, Kotlin, Jetpack Compose, SwiftUI

---

### Task 1: iOS Version Badge Implementation

**Files:**
- Modify: `ios/Airvoice/Views/HomeView.swift:270-275`

- [ ] **Step 1: Update HomeView.swift to append version text at bottomControls**

Modify [HomeView.swift](/ios/Airvoice/Views/HomeView.swift) around line 270-275 to add the version badge at the bottom of `bottomControls`:
```swift
            InputMethodTipsView(theme: theme)

            Text("v\(Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "0.3.1")")
                .font(.system(size: 11, weight: .medium, design: .rounded))
                .foregroundColor(theme.secondaryText.opacity(0.6))
                .frame(maxWidth: .infinity, alignment: .center)
                .padding(.top, 4)
        }
        .padding(.horizontal, 20)
        .padding(.bottom, 20)
```

- [ ] **Step 2: Commit iOS changes**

Run:
```bash
git add ios/Airvoice/Views/HomeView.swift
git commit -m "feat(ios): add version badge at the bottom of HomeView"
```

---

### Task 2: Android Version Badge Implementation

**Files:**
- Modify: `android/app/src/main/java/cc/yuler/airvoice/ui/screens/HomeScreen.kt:311-314`

- [ ] **Step 1: Update HomeScreen.kt to append version text at bottom controls**

Modify [HomeScreen.kt](/android/app/src/main/java/cc/yuler/airvoice/ui/screens/HomeScreen.kt) around line 311-314 to fetch the versionName dynamically and add the Text composable at the bottom:
```kotlin
                          }
                      }

                      InputMethodTipsView()

                      val context = LocalContext.current
                      val versionName = remember {
                          try {
                              context.packageManager.getPackageInfo(context.packageName, 0).versionName
                          } catch (e: Exception) {
                              "0.3.1"
                          }
                      }
                      Text(
                          text = "v$versionName",
                          fontSize = 11.sp,
                          fontWeight = FontWeight.Medium,
                          color = secondaryTextColor().copy(alpha = 0.6f),
                          modifier = Modifier.padding(top = 4.dp)
                      )
                  }
              }
          }
```

- [ ] **Step 2: Commit Android changes**

Run:
```bash
git add android/app/src/main/java/cc/yuler/airvoice/ui/screens/HomeScreen.kt
git commit -m "feat(android): add version badge at the bottom of HomeScreen"
```

---

### Task 3: Verification

- [ ] **Step 1: Run Android unit tests**

Run:
```bash
bash -c 'source ./scripts/java-env.sh && cd android && ./gradlew testDebugUnitTest'
```
Expected: All tests pass.
