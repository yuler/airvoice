package com.yule.airvoice.services

import android.content.Context
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.emptyPreferences
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import java.io.IOException
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.map

private val Context.dataStore by preferencesDataStore(name = "airvoice_prefs")

class StorageManager(private val context: Context) {
    companion object {
        private val KEY_WS = stringPreferencesKey("ws_url")
        private val KEY_TOKEN = stringPreferencesKey("token")
    }

    val wsUrlFlow: Flow<String?> = context.dataStore.data
        .catch { exception ->
            if (exception is IOException) {
                emit(emptyPreferences())
            } else {
                throw exception
            }
        }
        .map { prefs -> prefs[KEY_WS] }

    val tokenFlow: Flow<String?> = context.dataStore.data
        .catch { exception ->
            if (exception is IOException) {
                emit(emptyPreferences())
            } else {
                throw exception
            }
        }
        .map { prefs -> prefs[KEY_TOKEN] }

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
}
