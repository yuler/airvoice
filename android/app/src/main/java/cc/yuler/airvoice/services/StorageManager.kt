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
import kotlinx.coroutines.flow.map

private val Context.dataStore by preferencesDataStore(name = "airvoice_prefs")

private fun Flow<Preferences>.handleIOException(): Flow<Preferences> =
    catch { e -> if (e is IOException) emit(emptyPreferences()) else throw e }

data class ConnectionInfo(val wsUrl: String?, val token: String?)

class StorageManager(private val context: Context) {
    companion object {
        private val KEY_WS = stringPreferencesKey("ws_url")
        private val KEY_TOKEN = stringPreferencesKey("token")
        private val KEY_THEME = stringPreferencesKey("app_theme")
        private val KEY_HAS_SEEN_ONBOARDING = booleanPreferencesKey("has_seen_onboarding")
    }

    val connectionInfoFlow: Flow<ConnectionInfo> = context.dataStore.data
        .handleIOException()
        .map { prefs -> ConnectionInfo(prefs[KEY_WS], prefs[KEY_TOKEN]) }

    val themeFlow: Flow<String> = context.dataStore.data
        .handleIOException()
        .map { prefs -> prefs[KEY_THEME] ?: "light" }

    val hasSeenOnboardingFlow: Flow<Boolean> = context.dataStore.data
        .handleIOException()
        .map { prefs -> prefs[KEY_HAS_SEEN_ONBOARDING] ?: false }

    suspend fun saveConnection(wsUrl: String, token: String) {
        context.dataStore.edit { prefs ->
            prefs[KEY_WS] = wsUrl
            prefs[KEY_TOKEN] = token
        }
    }

    suspend fun clearConnection() {
        context.dataStore.edit { prefs ->
            prefs.remove(KEY_WS)
            prefs.remove(KEY_TOKEN)
        }
    }

    suspend fun saveTheme(theme: String) {
        context.dataStore.edit { prefs ->
            prefs[KEY_THEME] = theme
        }
    }

    suspend fun saveHasSeenOnboarding(completed: Boolean) {
        context.dataStore.edit { prefs ->
            prefs[KEY_HAS_SEEN_ONBOARDING] = completed
        }
    }
}
