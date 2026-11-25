# Project-specific ProGuard / R8 rules for Airvoice release builds.

# --- Kotlin ---
-keepattributes *Annotation*, InnerClasses, Signature, RuntimeVisibleAnnotations, AnnotationDefault

# kotlinx.serialization 1.6.2 bundles its own consumer rules; only keep app models.
-keep @kotlinx.serialization.Serializable class cc.yuler.airvoice.models.** { *; }

# --- OkHttp / Okio ---
-dontwarn okhttp3.**
-dontwarn okio.**

# --- Strip debug/verbose logs in release builds ---
-assumenosideeffects class android.util.Log {
    public static boolean isLoggable(java.lang.String, int);
    public static int v(...);
    public static int d(...);
}
