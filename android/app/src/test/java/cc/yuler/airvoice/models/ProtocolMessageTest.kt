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
        val msg = ProtocolMessage(type = "hello", id = "test-id", device = "Android", app = "0.1.0")
        val jsonStr = Json.encodeToString(ProtocolMessage.serializer(), msg)
        println("SERIALIZED JSON: $jsonStr")
        assert(jsonStr.contains("\"id\":\"test-id\""))
    }

    @Test
    fun testAckDecoding() {
        val jsonStr = """{"type":"ack","id":"msg-1","ok":true}"""
        val msg = Json { ignoreUnknownKeys = true }.decodeFromString<ProtocolMessage>(jsonStr)
        assertEquals("ack", msg.type)
        assertEquals("msg-1", msg.id)
        assertEquals(true, msg.ok)
    }
}
