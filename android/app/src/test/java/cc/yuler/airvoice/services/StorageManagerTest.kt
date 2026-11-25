package cc.yuler.airvoice.services

import android.content.Context
import androidx.datastore.preferences.core.edit
import androidx.test.core.app.ApplicationProvider
import cc.yuler.airvoice.models.AppLanguage
import cc.yuler.airvoice.models.AppSettings
import cc.yuler.airvoice.models.ThemeMode
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import javax.crypto.KeyGenerator

@RunWith(RobolectricTestRunner::class)
class StorageManagerTest {
    private lateinit var context: Context
    private lateinit var storageManager: StorageManager
    private lateinit var testCipher: TokenCipher

    @Before
    fun setUp() {
        runBlocking {
            context = ApplicationProvider.getApplicationContext()
            val key = KeyGenerator.getInstance("AES").apply { init(256) }.generateKey()
            testCipher = AesGcmTokenCipher { key }
            storageManager = StorageManager(context, testCipher)
            context.dataStore.edit { it.clear() }
        }
    }

    @After
    fun tearDown() {
        runBlocking {
            context.dataStore.edit { it.clear() }
        }
    }

    @Test
    fun testSaveEncryptsTokenAndClearRemovesAll() = runBlocking {
        var info = storageManager.connectionInfoFlow.first()
        assertNull(info.wsUrl)
        assertNull(info.token)
        assertFalse(info.allowRemote)

        storageManager.saveConnection("ws://192.168.1.5:7655/ws", "super-secret-token", allowRemote = true)

        // Raw DataStore must hold ciphertext, never plaintext token.
        val rawPrefs = context.dataStore.data.first()
        assertNull(rawPrefs[StorageManager.KEY_LEGACY_TOKEN])
        val storedCiphertext = rawPrefs[StorageManager.KEY_ENCRYPTED_TOKEN]
        assertNotNull(storedCiphertext)
        assertTrue(storedCiphertext!!.startsWith("v1:"))
        assertFalse(storedCiphertext.contains("super-secret-token"))

        info = storageManager.connectionInfoFlow.first()
        assertEquals("ws://192.168.1.5:7655/ws", info.wsUrl)
        assertEquals("super-secret-token", info.token)
        assertTrue(info.allowRemote)

        storageManager.clearConnection()
        info = storageManager.connectionInfoFlow.first()
        assertNull(info.wsUrl)
        assertNull(info.token)
        assertFalse(info.allowRemote)
    }

    @Test
    fun testLegacyPlaintextTokenMigratesToEncryptedStorage() = runBlocking {
        context.dataStore.edit { prefs ->
            prefs[StorageManager.KEY_WS] = "ws://192.168.1.10:7655/ws"
            prefs[StorageManager.KEY_LEGACY_TOKEN] = "legacy-plain-token"
        }

        val migrated = storageManager.readConnectionAndMigrateLegacy()
        assertEquals("ws://192.168.1.10:7655/ws", migrated.wsUrl)
        assertEquals("legacy-plain-token", migrated.token)

        val rawPrefs = context.dataStore.data.first()
        assertNull(rawPrefs[StorageManager.KEY_LEGACY_TOKEN])
        val encrypted = rawPrefs[StorageManager.KEY_ENCRYPTED_TOKEN]
        assertNotNull(encrypted)
        assertFalse(encrypted!!.contains("legacy-plain-token"))
        assertEquals("legacy-plain-token", storageManager.connectionInfoFlow.first().token)
    }

    @Test
    fun testCorruptedCiphertextReturnsNullToken() = runBlocking {
        context.dataStore.edit { prefs ->
            prefs[StorageManager.KEY_WS] = "ws://192.168.1.10:7655/ws"
            prefs[StorageManager.KEY_ENCRYPTED_TOKEN] = "v1:not-valid-ciphertext"
        }

        val info = storageManager.connectionInfoFlow.first()
        assertEquals("ws://192.168.1.10:7655/ws", info.wsUrl)
        assertNull(info.token)
    }

    @Test
    fun testLegacyThemeKeyReadsCleanlyViaSettingsFlow() = runBlocking {
        assertEquals(ThemeMode.SYSTEM, storageManager.settingsFlow.first().themeMode)
        context.dataStore.edit { prefs -> prefs[StorageManager.KEY_THEME] = "dark" }
        assertEquals(ThemeMode.DARK, storageManager.settingsFlow.first().themeMode)
    }

    @Test
    fun testSettingsPersistenceAndResetPreservesConnection() = runBlocking {
        assertEquals(AppSettings(), storageManager.settingsFlow.first())

        storageManager.saveConnection("ws://192.168.1.5:7655/ws", "keep-token", allowRemote = false)
        val updated = AppSettings(
            themeMode = ThemeMode.DARK,
            language = AppLanguage.EN,
            autoSendEnabled = false,
            autoSendDelayMs = 2500L,
            hapticFeedbackEnabled = false,
            keepScreenAwake = false,
            showInputTips = false,
            sendTimeoutMs = 10000L
        )
        storageManager.saveSettings(updated)

        val saved = storageManager.settingsFlow.first()
        assertEquals(updated, saved)

        storageManager.resetSettings()
        assertEquals(AppSettings(), storageManager.settingsFlow.first())

        val conn = storageManager.connectionInfoFlow.first()
        assertEquals("ws://192.168.1.5:7655/ws", conn.wsUrl)
        assertEquals("keep-token", conn.token)
    }
}
