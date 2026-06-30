package cc.yuler.airvoice.services

import android.content.Context
import androidx.test.core.app.ApplicationProvider
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@RunWith(RobolectricTestRunner::class)
class StorageManagerTest {
    private lateinit var context: Context
    private lateinit var storageManager: StorageManager

    @Before
    fun setUp() {
        context = ApplicationProvider.getApplicationContext()
        storageManager = StorageManager(context)
    }

    @Test
    fun testSaveAndClearConnection() = runBlocking {
        var info = storageManager.connectionInfoFlow.first()
        assertNull(info.wsUrl)
        assertNull(info.token)

        storageManager.saveConnection("ws://test", "tok")
        info = storageManager.connectionInfoFlow.first()
        assertEquals("ws://test", info.wsUrl)
        assertEquals("tok", info.token)

        storageManager.clearConnection()
        info = storageManager.connectionInfoFlow.first()
        assertNull(info.wsUrl)
        assertNull(info.token)
    }

    @Test
    fun testSaveTheme() = runBlocking {
        assertEquals("light", storageManager.themeFlow.first())
        storageManager.saveTheme("dark")
        assertEquals("dark", storageManager.themeFlow.first())
    }

    @Test
    fun testSaveOnboarding() = runBlocking {
        assertFalse(storageManager.hasSeenOnboardingFlow.first())
        storageManager.saveHasSeenOnboarding(true)
        assertTrue(storageManager.hasSeenOnboardingFlow.first())
    }
}
