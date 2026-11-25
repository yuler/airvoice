package cc.yuler.airvoice.models

import cc.yuler.airvoice.ui.i18n.EnStrings
import cc.yuler.airvoice.ui.i18n.ZhStrings
import cc.yuler.airvoice.ui.i18n.stringsFor
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import java.util.Locale

class AppSettingsTest {
    @Test
    fun testThemeModeFromStorageDefaultsToSystemAndPreservesLegacyValues() {
        assertEquals(ThemeMode.SYSTEM, ThemeMode.fromStorage(null))
        assertEquals(ThemeMode.SYSTEM, ThemeMode.fromStorage("unknown"))
        assertEquals(ThemeMode.SYSTEM, ThemeMode.fromStorage("system"))
        assertEquals(ThemeMode.LIGHT, ThemeMode.fromStorage("light"))
        assertEquals(ThemeMode.DARK, ThemeMode.fromStorage("dark"))
    }

    @Test
    fun testAppLanguageResolution() {
        assertTrue(AppLanguage.ZH.isChinese(Locale.US))
        assertFalse(AppLanguage.EN.isChinese(Locale.CHINA))
        assertTrue(AppLanguage.SYSTEM.isChinese(Locale.SIMPLIFIED_CHINESE))
        assertTrue(AppLanguage.SYSTEM.isChinese(Locale.TRADITIONAL_CHINESE))
        assertFalse(AppLanguage.SYSTEM.isChinese(Locale.US))
        assertFalse(AppLanguage.SYSTEM.isChinese(Locale.JAPAN))

        assertEquals(ZhStrings, stringsFor(AppLanguage.ZH, Locale.US))
        assertEquals(EnStrings, stringsFor(AppLanguage.EN, Locale.CHINA))
        assertEquals(ZhStrings, stringsFor(AppLanguage.SYSTEM, Locale.CHINA))
        assertEquals(EnStrings, stringsFor(AppLanguage.SYSTEM, Locale.US))
    }

    @Test
    fun testSanitizeDelayAndTimeoutOptions() {
        assertEquals(800L, AppSettings.sanitizeDelayMs(800L))
        assertEquals(4000L, AppSettings.sanitizeDelayMs(4000L))
        assertEquals(AppSettings.DEFAULT_AUTO_SEND_DELAY_MS, AppSettings.sanitizeDelayMs(999L))
        assertEquals(AppSettings.DEFAULT_AUTO_SEND_DELAY_MS, AppSettings.sanitizeDelayMs(null))

        assertEquals(3000L, AppSettings.sanitizeTimeoutMs(3000L))
        assertEquals(10000L, AppSettings.sanitizeTimeoutMs(10000L))
        assertEquals(AppSettings.DEFAULT_SEND_TIMEOUT_MS, AppSettings.sanitizeTimeoutMs(1234L))
        assertEquals(AppSettings.DEFAULT_SEND_TIMEOUT_MS, AppSettings.sanitizeTimeoutMs(null))
    }
}
