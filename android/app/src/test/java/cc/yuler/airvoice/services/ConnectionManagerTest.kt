package cc.yuler.airvoice.services

import cc.yuler.airvoice.models.ProtocolMessage
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.json.Json
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
        server.shutdown()
    }

    @Test
    fun testConnectionManagerInitialState() = runBlocking {
        assertEquals(ConnectionStatus.Disconnected, connectionManager.status.value)
    }
}
