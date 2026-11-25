package cc.yuler.airvoice.ui.viewmodel

import android.app.Application
import androidx.datastore.preferences.core.edit
import androidx.test.core.app.ApplicationProvider
import cc.yuler.airvoice.models.AppLanguage
import cc.yuler.airvoice.models.AppSettings
import cc.yuler.airvoice.models.PairingPayload
import cc.yuler.airvoice.models.ThemeMode
import cc.yuler.airvoice.services.AesGcmTokenCipher
import cc.yuler.airvoice.services.ConnectionManager
import cc.yuler.airvoice.services.ConnectionStatus
import cc.yuler.airvoice.services.SendTrigger
import cc.yuler.airvoice.services.StorageManager
import cc.yuler.airvoice.services.dataStore
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.async
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withTimeout
import okhttp3.OkHttpClient
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import java.util.concurrent.atomic.AtomicInteger
import javax.crypto.KeyGenerator

@RunWith(RobolectricTestRunner::class)
class AirvoiceViewModelTest {
    private lateinit var application: Application
    private lateinit var storageManager: StorageManager
    private lateinit var connectionManager: ConnectionManager
    private lateinit var viewModel: AirvoiceViewModel

    @Before
    fun setUp() {
        runBlocking {
            application = ApplicationProvider.getApplicationContext()
            application.dataStore.edit { it.clear() }
            val key = KeyGenerator.getInstance("AES").apply { init(256) }.generateKey()
            storageManager = StorageManager(application, AesGcmTokenCipher { key })
            connectionManager = ConnectionManager(OkHttpClient())
            viewModel = AirvoiceViewModel(application, storageManager, connectionManager)
        }
    }

    @After
    fun tearDown() {
        runBlocking {
            connectionManager.disconnect()
            viewModel.autoSendController.cleanup()
            application.dataStore.edit { it.clear() }
        }
    }

    @Test
    fun testCancelAutoSendDoesNotRetryOrShowFailureToast() = runBlocking {
        val server = MockWebServer()
        val connected = CompletableDeferred<Unit>()
        val firstTextReceived = CompletableDeferred<Unit>()
        val textFrameCount = AtomicInteger(0)

        server.enqueue(
            MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
                override fun onOpen(webSocket: WebSocket, response: Response) {
                    webSocket.send("""{"type":"hello","host":"test-pc","version":"0.1.0"}""")
                    connected.complete(Unit)
                }

                override fun onMessage(webSocket: WebSocket, text: String) {
                    if (text.contains(""""type":"text"""")) {
                        textFrameCount.incrementAndGet()
                        firstTextReceived.complete(Unit)
                    }
                }
            })
        )
        server.start()
        try {
            val wsUrl = server.url("/ws").toString().replaceFirst("http://", "ws://")
            connectionManager.connect(wsUrl, "tok-1")
            withTimeout(3000L) {
                connected.await()
                connectionManager.status.first { it is ConnectionStatus.Connected }
            }

            val sendJob = async {
                viewModel.autoSendController.attemptSend("auto send to cancel", SendTrigger.AUTO)
            }
            withTimeout(3000L) { firstTextReceived.await() }
            assertEquals(1, textFrameCount.get())

            viewModel.cancelSend()
            assertFalse(sendJob.await())
            assertEquals(viewModel.currentStrings().toastSendCancelled, viewModel.toastMessage.value)
            assertFalse(viewModel.isToastError.value)
            assertFalse(viewModel.sendTimedOut.value)

            // Wait past the 400ms auto-retry window to confirm NO automatic resend occurs.
            delay(700L)
            assertEquals(1, textFrameCount.get())
            assertEquals(viewModel.currentStrings().toastSendCancelled, viewModel.toastMessage.value)
            assertFalse(viewModel.isToastError.value)
            assertFalse(viewModel.sendTimedOut.value)
        } finally {
            server.shutdown()
        }
    }

    @Test
    fun testPairAndConnectRejectsCleartextRemoteAndUnconfirmedWss() = runBlocking {
        viewModel.pairAndConnect(PairingPayload(v = 1, ws = "ws://evil.example.com/ws", token = "tok"))
        assertTrue(viewModel.isToastError.value)
        assertNull(storageManager.connectionInfoFlow.first().wsUrl)
        assertEquals(ConnectionStatus.Disconnected, connectionManager.status.value)

        viewModel.pairAndConnect(
            PairingPayload(v = 1, ws = "wss://remote.example.com/ws", token = "tok"),
            allowRemote = false
        )
        assertTrue(viewModel.isToastError.value)
        assertNull(storageManager.connectionInfoFlow.first().wsUrl)
        assertEquals(ConnectionStatus.Disconnected, connectionManager.status.value)
    }

    @Test
    fun testSettingsMutationsAndLanguageSwitching() = runBlocking {
        viewModel.setThemeMode(ThemeMode.DARK)
        viewModel.setLanguage(AppLanguage.EN)
        viewModel.setAutoSendEnabled(false)
        viewModel.setAutoSendDelayMs(2500L)
        viewModel.setHapticFeedbackEnabled(false)
        viewModel.setKeepScreenAwake(false)
        viewModel.setShowInputTips(false)
        viewModel.setSendTimeoutMs(10000L)

        val current = viewModel.settings.value
        assertEquals(ThemeMode.DARK, current.themeMode)
        assertEquals(AppLanguage.EN, current.language)
        assertFalse(current.autoSendEnabled)
        assertEquals(2500L, current.autoSendDelayMs)
        assertFalse(current.hapticFeedbackEnabled)
        assertFalse(current.keepScreenAwake)
        assertFalse(current.showInputTips)
        assertEquals(10000L, current.sendTimeoutMs)
        assertEquals("Send cancelled", viewModel.currentStrings().toastSendCancelled)

        viewModel.setLanguage(AppLanguage.ZH)
        assertEquals("已取消发送", viewModel.currentStrings().toastSendCancelled)

        viewModel.resetSettings()
        assertEquals(AppSettings(), viewModel.settings.value)
    }
}
