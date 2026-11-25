package cc.yuler.airvoice.services

import android.content.Context
import androidx.datastore.preferences.core.MutablePreferences
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.emptyPreferences
import androidx.datastore.preferences.core.longPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import cc.yuler.airvoice.models.AppLanguage
import cc.yuler.airvoice.models.AppSettings
import cc.yuler.airvoice.models.ThemeMode
import java.io.IOException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.withContext

internal val Context.dataStore by preferencesDataStore(name = "airvoice_prefs")

private fun Flow<Preferences>.handleIOException(): Flow<Preferences> =
    catch { e -> if (e is IOException) emit(emptyPreferences()) else throw e }

data class ConnectionInfo(
    val wsUrl: String?,
    val token: String?,
    val allowRemote: Boolean = false
)

class StorageManager(
    private val context: Context,
    private val tokenCipher: TokenCipher = AndroidKeystoreTokenCipher()
) {
    companion object {
        internal val KEY_WS = stringPreferencesKey("ws_url")
        internal val KEY_LEGACY_TOKEN = stringPreferencesKey("token")
        internal val KEY_ENCRYPTED_TOKEN = stringPreferencesKey("token_enc")
        internal val KEY_ALLOW_REMOTE = booleanPreferencesKey("allow_remote")
        internal val KEY_THEME = stringPreferencesKey("app_theme")
        internal val KEY_LANGUAGE = stringPreferencesKey("app_language")
        internal val KEY_AUTO_SEND_ENABLED = booleanPreferencesKey("auto_send_enabled")
        internal val KEY_AUTO_SEND_DELAY_MS = longPreferencesKey("auto_send_delay_ms")
        internal val KEY_HAPTIC_FEEDBACK = booleanPreferencesKey("haptic_feedback_enabled")
        internal val KEY_KEEP_SCREEN_AWAKE = booleanPreferencesKey("keep_screen_awake")
        internal val KEY_SHOW_INPUT_TIPS = booleanPreferencesKey("show_input_tips")
        internal val KEY_SEND_TIMEOUT_MS = longPreferencesKey("send_timeout_ms")
    }

    val connectionInfoFlow: Flow<ConnectionInfo> = context.dataStore.data
        .handleIOException()
        .map { prefs -> decodeConnectionInfo(prefs) }

    val settingsFlow: Flow<AppSettings> = context.dataStore.data
        .handleIOException()
        .map { prefs -> decodeSettings(prefs) }

    private fun decodeSettings(prefs: Preferences): AppSettings = AppSettings(
        themeMode = ThemeMode.fromStorage(prefs[KEY_THEME]),
        language = AppLanguage.fromStorage(prefs[KEY_LANGUAGE]),
        autoSendEnabled = prefs[KEY_AUTO_SEND_ENABLED] ?: true,
        autoSendDelayMs = AppSettings.sanitizeDelayMs(prefs[KEY_AUTO_SEND_DELAY_MS]),
        hapticFeedbackEnabled = prefs[KEY_HAPTIC_FEEDBACK] ?: true,
        keepScreenAwake = prefs[KEY_KEEP_SCREEN_AWAKE] ?: true,
        showInputTips = prefs[KEY_SHOW_INPUT_TIPS] ?: true,
        sendTimeoutMs = AppSettings.sanitizeTimeoutMs(prefs[KEY_SEND_TIMEOUT_MS])
    )

    private fun decodeConnectionInfo(prefs: Preferences): ConnectionInfo {
        val wsUrl = prefs[KEY_WS]
        val encrypted = prefs[KEY_ENCRYPTED_TOKEN]
        val token = when {
            !encrypted.isNullOrEmpty() -> tokenCipher.decrypt(encrypted)
            else -> prefs[KEY_LEGACY_TOKEN]
        }
        val allowRemote = prefs[KEY_ALLOW_REMOTE] ?: false
        return ConnectionInfo(wsUrl = wsUrl, token = token, allowRemote = allowRemote)
    }

    private suspend fun editPrefs(block: (MutablePreferences) -> Unit) {
        withContext(Dispatchers.IO) {
            context.dataStore.edit(block)
        }
    }

    /**
     * Reads stored connection credentials and migrates any legacy plaintext token
     * into Keystore-encrypted storage, purging the plaintext preference key.
     */
    suspend fun readConnectionAndMigrateLegacy(): ConnectionInfo = withContext(Dispatchers.IO) {
        val prefs = context.dataStore.data.handleIOException().first()
        val legacyToken = prefs[KEY_LEGACY_TOKEN]
        val encryptedToken = prefs[KEY_ENCRYPTED_TOKEN]

        if (!legacyToken.isNullOrEmpty() && encryptedToken.isNullOrEmpty()) {
            val encrypted = tokenCipher.encrypt(legacyToken)
            context.dataStore.edit { editable ->
                editable[KEY_ENCRYPTED_TOKEN] = encrypted
                editable.remove(KEY_LEGACY_TOKEN)
            }
            return@withContext ConnectionInfo(
                wsUrl = prefs[KEY_WS],
                token = legacyToken,
                allowRemote = prefs[KEY_ALLOW_REMOTE] ?: false
            )
        } else if (legacyToken != null) {
            context.dataStore.edit { editable ->
                editable.remove(KEY_LEGACY_TOKEN)
            }
        }

        decodeConnectionInfo(prefs)
    }

    suspend fun saveConnection(wsUrl: String, token: String, allowRemote: Boolean = false) {
        editPrefs { prefs ->
            prefs[KEY_WS] = wsUrl
            prefs[KEY_ENCRYPTED_TOKEN] = tokenCipher.encrypt(token)
            prefs[KEY_ALLOW_REMOTE] = allowRemote
            prefs.remove(KEY_LEGACY_TOKEN)
        }
    }

    suspend fun clearConnection() {
        editPrefs { prefs ->
            prefs.remove(KEY_WS)
            prefs.remove(KEY_ENCRYPTED_TOKEN)
            prefs.remove(KEY_LEGACY_TOKEN)
            prefs.remove(KEY_ALLOW_REMOTE)
        }
    }

    suspend fun saveSettings(settings: AppSettings) {
        editPrefs { prefs ->
            prefs[KEY_THEME] = settings.themeMode.storageValue
            prefs[KEY_LANGUAGE] = settings.language.storageValue
            prefs[KEY_AUTO_SEND_ENABLED] = settings.autoSendEnabled
            prefs[KEY_AUTO_SEND_DELAY_MS] = AppSettings.sanitizeDelayMs(settings.autoSendDelayMs)
            prefs[KEY_HAPTIC_FEEDBACK] = settings.hapticFeedbackEnabled
            prefs[KEY_KEEP_SCREEN_AWAKE] = settings.keepScreenAwake
            prefs[KEY_SHOW_INPUT_TIPS] = settings.showInputTips
            prefs[KEY_SEND_TIMEOUT_MS] = AppSettings.sanitizeTimeoutMs(settings.sendTimeoutMs)
        }
    }

    /**
     * Resets all user preferences back to defaults while preserving paired connection credentials.
     */
    suspend fun resetSettings() {
        editPrefs { prefs ->
            prefs.remove(KEY_THEME)
            prefs.remove(KEY_LANGUAGE)
            prefs.remove(KEY_AUTO_SEND_ENABLED)
            prefs.remove(KEY_AUTO_SEND_DELAY_MS)
            prefs.remove(KEY_HAPTIC_FEEDBACK)
            prefs.remove(KEY_KEEP_SCREEN_AWAKE)
            prefs.remove(KEY_SHOW_INPUT_TIPS)
            prefs.remove(KEY_SEND_TIMEOUT_MS)
        }
    }
}
