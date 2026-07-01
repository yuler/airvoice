# Project-specific ProGuard / R8 rules for Airvoice release builds.

# --- Kotlin ---
-keepattributes *Annotation*, InnerClasses, Signature, RuntimeVisibleAnnotations, AnnotationDefault

# kotlinx.serialization 1.6.2 bundles its own consumer rules; only keep app models.
-keep @kotlinx.serialization.Serializable class cc.yuler.airvoice.models.** { *; }

# --- OkHttp / Okio ---
-dontwarn okhttp3.**
-dontwarn okio.**
