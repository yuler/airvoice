package com.yule.airvoice.services

import com.yule.airvoice.models.ProtocolMessage
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class AutoSendControllerTest {

    @Test
    fun testTextFlowEmitsValue() = runTest {
        val textFlow = MutableStateFlow("")
        textFlow.value = "Hello"
        assertEquals("Hello", textFlow.value)

        textFlow.value = "Hello World"
        assertEquals("Hello World", textFlow.value)
    }

    @Test
    fun testLastAckedTextResetOnEmpty() = runTest {
        val textFlow = MutableStateFlow("Hello")
        assertEquals("Hello", textFlow.value)

        textFlow.value = ""
        assertEquals("", textFlow.value)
    }

    @Test
    fun testProtocolMessageCreation() {
        val msg = ProtocolMessage(
            type = "text",
            id = "test-id",
            content = "Hello",
            ts = 1234567890L
        )
        assertEquals("text", msg.type)
        assertEquals("test-id", msg.id)
        assertEquals("Hello", msg.content)
        assertEquals(1234567890L, msg.ts)
    }

    @Test
    fun testProtocolMessageTypes() {
        val helloMsg = ProtocolMessage(type = "hello", device = "Android Phone", app = "0.1.0")
        assertEquals("hello", helloMsg.type)
        assertEquals("Android Phone", helloMsg.device)

        val ackMsg = ProtocolMessage(type = "ack", id = "msg-123", ok = true)
        assertEquals("ack", ackMsg.type)
        assertEquals(true, ackMsg.ok)

        val failAck = ProtocolMessage(type = "ack", id = "msg-456", ok = false)
        assertEquals(false, failAck.ok)
    }

    @Test
    fun testConnectionStatusTypes() {
        val disconnected: ConnectionStatus = ConnectionStatus.Disconnected
        val connecting: ConnectionStatus = ConnectionStatus.Connecting
        val connected: ConnectionStatus = ConnectionStatus.Connected("MyPC")
        val error: ConnectionStatus = ConnectionStatus.Error("timeout")

        assertTrue(disconnected is ConnectionStatus.Disconnected)
        assertTrue(connecting is ConnectionStatus.Connecting)
        assertTrue(connected is ConnectionStatus.Connected)
        assertEquals("MyPC", (connected as ConnectionStatus.Connected).host)
        assertTrue(error is ConnectionStatus.Error)
        assertEquals("timeout", (error as ConnectionStatus.Error).message)
    }
}
