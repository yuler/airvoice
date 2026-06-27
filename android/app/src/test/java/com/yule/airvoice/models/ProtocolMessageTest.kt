package com.yule.airvoice.models

import kotlinx.serialization.json.Json
import org.junit.Assert.assertEquals
import org.junit.Test

class ProtocolMessageTest {
    @Test
    fun testPairingPayloadDeserialization() {
        val jsonStr = """{"v":1,"ws":"ws://192.168.1.10:7383/ws","token":"test-token"}"""
        val payload = Json.decodeFromString<PairingPayload>(jsonStr)
        assertEquals(1, payload.v)
        assertEquals("ws://192.168.1.10:7383/ws", payload.ws)
        assertEquals("test-token", payload.token)
    }

    @Test
    fun testProtocolMessageSerialization() {
        val msg = ProtocolMessage(type = "hello", device = "Android", app = "0.1.0")
        val jsonStr = Json.encodeToString(ProtocolMessage.serializer(), msg)
        // We assert it contains type hello
        assert(jsonStr.contains("\"type\":\"hello\""))
    }
}
