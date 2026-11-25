package cc.yuler.airvoice.services

import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.async
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.runBlocking
import okhttp3.OkHttpClient
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import java.lang.reflect.Field
import java.util.concurrent.CopyOnWriteArrayList

@RunWith(RobolectricTestRunner::class)
class AutoSendControllerTest {
    private lateinit var connectionManager: ConnectionManager
    private lateinit var textFlow: MutableStateFlow<String>

    @Before
    fun setUp() {
        connectionManager = ConnectionManager(OkHttpClient())
        textFlow = MutableStateFlow("")
    }

    @After
    fun tearDown() {
        connectionManager.disconnect()
    }

    @Test
    fun testCountdownStateTransitions() = runBlocking {
        val controller = AutoSendController(textFlow, connectionManager) { _, _, _ -> }

        assertFalse(controller.countdownActive.value)
        assertEquals(0, controller.countdownToken.value)

        controller.textDidChange("hello")
        assertTrue(controller.countdownActive.value)
        assertEquals(1, controller.countdownToken.value)

        controller.textDidChange("hello world")
        assertTrue(controller.countdownActive.value)
        assertEquals(2, controller.countdownToken.value)

        controller.cleanup()
    }

    @Test
    fun testCountdownStopsWhileInFlight() = runBlocking {
        val controller = AutoSendController(textFlow, connectionManager) { _, _, _ -> }

        controller.textDidChange("hello")
        assertTrue(controller.countdownActive.value)

        setInFlight(controller, true)
        controller.textDidChange("hello world")

        assertFalse(controller.countdownActive.value)
        assertEquals(1, controller.countdownToken.value)

        controller.cleanup()
    }

    @Test
    fun testResetLastAckedWhenTextCleared() {
        val controller = AutoSendController(textFlow, connectionManager) { _, _, _ -> }
        setLastAckedText(controller, "hello")

        controller.textDidChange("")

        assertEquals("", getLastAckedText(controller))
        controller.cleanup()
    }

    @Test
    fun testClearInFlightCompletesWithCancelledInsteadOfFailed() = runBlocking {
        val server = MockWebServer()
        val opened = CompletableDeferred<Unit>()
        val receivedText = CompletableDeferred<String>()
        server.enqueue(
            MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
                override fun onOpen(webSocket: WebSocket, response: Response) {
                    webSocket.send("""{"type":"hello","host":"test-pc","version":"0.1.0"}""")
                    opened.complete(Unit)
                }

                override fun onMessage(webSocket: WebSocket, text: String) {
                    if (text.contains(""""type":"text"""")) {
                        receivedText.complete(text)
                    }
                }
            })
        )
        server.start()
        try {
            val wsUrl = server.url("/ws").toString().replaceFirst("http://", "ws://")
            connectionManager.connect(wsUrl, "valid-token")
            opened.await()

            val results = CopyOnWriteArrayList<Triple<SendResult, String, SendTrigger>>()
            val controller = AutoSendController(textFlow, connectionManager) { result, text, trigger ->
                results.add(Triple(result, text, trigger))
            }

            val sendDeferred = async {
                controller.attemptSend("cancel me", SendTrigger.AUTO)
            }

            // Wait until the frame is actually sent over the socket and in-flight.
            receivedText.await()
            assertTrue(controller.inFlight.value)

            controller.clearInFlight()
            assertFalse(controller.inFlight.value)

            val succeeded = sendDeferred.await()
            assertFalse(succeeded)
            assertEquals(1, results.size)
            assertEquals(SendResult.CANCELLED, results[0].first)
            assertEquals("cancel me", results[0].second)
            assertEquals(SendTrigger.AUTO, results[0].third)

            controller.cleanup()
        } finally {
            connectionManager.disconnect()
            server.shutdown()
        }
    }

    private fun setInFlight(controller: AutoSendController, value: Boolean) {
        val field: Field = AutoSendController::class.java.getDeclaredField("_inFlight")
        field.isAccessible = true
        @Suppress("UNCHECKED_CAST")
        val flow = field.get(controller) as MutableStateFlow<Boolean>
        flow.value = value
    }

    private fun setLastAckedText(controller: AutoSendController, value: String) {
        val field: Field = AutoSendController::class.java.getDeclaredField("lastAckedText")
        field.isAccessible = true
        field.set(controller, value)
    }

    private fun getLastAckedText(controller: AutoSendController): String {
        val field: Field = AutoSendController::class.java.getDeclaredField("lastAckedText")
        field.isAccessible = true
        return field.get(controller) as String
    }
}
