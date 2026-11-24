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
import org.junit.Before
import org.junit.Test

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
    fun testUnauthorized401StopsReconnectAndEmitsAuthExpired() = runBlocking {
        server.enqueue(MockResponse().setResponseCode(401).setBody("Unauthorized"))
        val wsUrl = server.url("/ws").toString().replaceFirst("http://", "ws://")

        val authExpired = async(start = CoroutineStart.UNDISPATCHED) {
            connectionManager.authExpiredEvents.first()
        }

        connectionManager.connect(wsUrl, "stale-token")

        withTimeout(3000L) {
            authExpired.await()
        }

        assertEquals(ConnectionStatus.Disconnected, connectionManager.status.value)
        assertEquals(1, server.requestCount)

        // Ensure no automatic retry is triggered after the default 2s initial backoff.
        delay(2300L)
        assertEquals(1, server.requestCount)
    }
}
