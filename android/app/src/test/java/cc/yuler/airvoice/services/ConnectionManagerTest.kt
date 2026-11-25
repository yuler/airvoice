package cc.yuler.airvoice.services

import kotlinx.coroutines.CoroutineStart
import kotlinx.coroutines.async
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withTimeout
import okhttp3.OkHttpClient
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import java.util.concurrent.TimeUnit

class ConnectionManagerTest {
    private val server = MockWebServer()
    private lateinit var connectionManager: ConnectionManager

    @Before
    fun setUp() {
        server.start()
        connectionManager = ConnectionManager(OkHttpClient())
    }

    @After
    fun tearDown() {
        connectionManager.disconnect()
        server.shutdown()
    }

    @Test
    fun testConnectionManagerInitialState() = runBlocking {
        assertEquals(ConnectionStatus.Disconnected, connectionManager.status.value)
    }

    @Test
    fun testUnauthorized401StopsReconnectAndEmitsAuthExpiredAndUsesBearerHeader() = runBlocking {
        server.enqueue(MockResponse().setResponseCode(401).setBody("Unauthorized"))
        val wsUrl = server.url("/ws").toString().replaceFirst("http://", "ws://")

        val authExpired = async(start = CoroutineStart.UNDISPATCHED) {
            connectionManager.authExpiredEvents.first()
        }

        connectionManager.connect(wsUrl, "stale-token")

        withTimeout(3000L) {
            authExpired.await()
        }

        val recorded = server.takeRequest(2, TimeUnit.SECONDS)
        assertEquals("Bearer stale-token", recorded?.getHeader("Authorization"))
        assertNull(recorded?.requestUrl?.queryParameter("token"))

        assertEquals(ConnectionStatus.Disconnected, connectionManager.status.value)
        assertEquals(1, server.requestCount)

        // Ensure no automatic retry is triggered after the default 2s initial backoff.
        delay(2300L)
        assertEquals(1, server.requestCount)
    }

    @Test
    fun testInvalidAndRemoteUrlsRejectedWithoutLeakingTokenInErrorStatus() {
        val secretToken = "super-secret-token-123"

        connectionManager.connect("ws://evil.example.com/ws", secretToken)
        val err1 = connectionManager.status.value
        assertTrue(err1 is ConnectionStatus.Error)
        assertFalse((err1 as ConnectionStatus.Error).message.contains(secretToken))
        assertFalse(err1.message.contains("evil.example.com"))

        connectionManager.connect("not-a-url?token=$secretToken", secretToken)
        val err2 = connectionManager.status.value
        assertTrue(err2 is ConnectionStatus.Error)
        assertFalse((err2 as ConnectionStatus.Error).message.contains(secretToken))

        connectionManager.connect("wss://remote.example.com/ws", secretToken, allowRemote = false)
        val err3 = connectionManager.status.value
        assertTrue(err3 is ConnectionStatus.Error)
        assertFalse((err3 as ConnectionStatus.Error).message.contains(secretToken))
    }
}
