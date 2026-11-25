package cc.yuler.airvoice.services

import android.content.Context
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.emptyPreferences
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import java.io.IOException
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map

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
    }

    val connectionInfoFlow: Flow<ConnectionInfo> = context.dataStore.data
        .handleIOException()
        .map { prefs -> decodeConnectionInfo(prefs) }

    val themeFlow: Flow<String> = context.dataStore.data
        .handleIOException()
        .map { prefs -> prefs[KEY_THEME] ?: "light" }

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

    /**
     * Reads stored connection credentials and migrates any legacy plaintext token
     * into Keystore-encrypted storage, purging the plaintext preference key.
     */
    suspend fun readConnectionAndMigrateLegacy(): ConnectionInfo {
        val prefs = context.dataStore.data.handleIOException().first()
        val legacyToken = prefs[KEY_LEGACY_TOKEN]
        val encryptedToken = prefs[KEY_ENCRYPTED_TOKEN]

        if (!legacyToken.isNullOrEmpty() && encryptedToken.isNullOrEmpty()) {
            val encrypted = tokenCipher.encrypt(legacyToken)
            context.dataStore.edit { editable ->
                editable[KEY_ENCRYPTED_TOKEN] = encrypted
                editable.remove(KEY_LEGACY_TOKEN)
            }
            return ConnectionInfo(
                wsUrl = prefs[KEY_WS],
                token = legacyToken,
                allowRemote = prefs[KEY_ALLOW_REMOTE] ?: false
            )
        } else if (legacyToken != null) {
            context.dataStore.edit { editable ->
                editable.remove(KEY_LEGACY_TOKEN)
            }
        }

        return decodeConnectionInfo(prefs)
    }

    suspend fun saveConnection(wsUrl: String, token: String, allowRemote: Boolean = false) {
        val encryptedToken = tokenCipher.encrypt(token)
        context.dataStore.edit { prefs ->
            prefs[KEY_WS] = wsUrl
            prefs[KEY_ENCRYPTED_TOKEN] = encryptedToken
            prefs[KEY_ALLOW_REMOTE] = allowRemote
            prefs.remove(KEY_LEGACY_TOKEN)
        }
    }

    suspend fun clearConnection() {
        context.dataStore.edit { prefs ->
            prefs.remove(KEY_WS)
            prefs.remove(KEY_ENCRYPTED_TOKEN)
            prefs.remove(KEY_LEGACY_TOKEN)
            prefs.remove(KEY_ALLOW_REMOTE)
        }
    }

    suspend fun saveTheme(theme: String) {
        context.dataStore.edit { prefs ->
            prefs[KEY_THEME] = theme
        }
    }
}
