package cc.yuler.airvoice.models

import java.util.Locale

enum class ThemeMode(val storageValue: String) {
    SYSTEM("system"),
    LIGHT("light"),
    DARK("dark");

    companion object {
        fun fromStorage(value: String?): ThemeMode =
            entries.firstOrNull { it.storageValue == value } ?: SYSTEM
    }
}

enum class AppLanguage(val storageValue: String) {
    SYSTEM("system"),
    ZH("zh"),
    EN("en");

    fun isChinese(locale: Locale = Locale.getDefault()): Boolean = when (this) {
        ZH -> true
        EN -> false
        SYSTEM -> locale.language.equals("zh", ignoreCase = true)
    }

    companion object {
        fun fromStorage(value: String?): AppLanguage =
            entries.firstOrNull { it.storageValue == value } ?: SYSTEM
    }
}

data class AppSettings(
    val themeMode: ThemeMode = ThemeMode.SYSTEM,
    val language: AppLanguage = AppLanguage.SYSTEM,
    val autoSendEnabled: Boolean = true,
    val autoSendDelayMs: Long = DEFAULT_AUTO_SEND_DELAY_MS,
    val hapticFeedbackEnabled: Boolean = true,
    val keepScreenAwake: Boolean = true,
    val showInputTips: Boolean = true,
    val sendTimeoutMs: Long = DEFAULT_SEND_TIMEOUT_MS
) {
    companion object {
        const val DEFAULT_AUTO_SEND_DELAY_MS = 1500L
        const val DEFAULT_SEND_TIMEOUT_MS = 5000L

        val AUTO_SEND_DELAY_OPTIONS = listOf(800L, 1500L, 2500L, 4000L)
        val SEND_TIMEOUT_OPTIONS = listOf(3000L, 5000L, 10000L)

        fun sanitizeDelayMs(value: Long?): Long =
            if (value != null && value in AUTO_SEND_DELAY_OPTIONS) value else DEFAULT_AUTO_SEND_DELAY_MS

        fun sanitizeTimeoutMs(value: Long?): Long =
            if (value != null && value in SEND_TIMEOUT_OPTIONS) value else DEFAULT_SEND_TIMEOUT_MS
    }
}
