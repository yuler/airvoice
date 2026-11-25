package cc.yuler.airvoice.models

import cc.yuler.airvoice.AirvoiceApplication
import okhttp3.Request
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Test
import java.io.IOException
import java.net.InetAddress

class PairingValidatorTest {
    @Test
    fun testValidLocalAddressesAccepted() {
        val localUrls = listOf(
            "ws://192.168.1.100:7383/ws",
            "ws://10.0.0.5:7655/ws",
            "ws://172.16.0.2:7655/ws",
            "ws://172.31.255.1:7655/ws",
            "ws://127.0.0.1:7655/ws",
            "ws://169.254.1.2:7655/ws",
            "ws://100.64.0.1:7655/ws",
            "ws://100.127.255.254:7655/ws",
            "ws://localhost:7655/ws",
            "ws://my-mac.local:7655/ws",
            "ws://[::1]:7655/ws",
            "ws://[fd7a:115c:a1e0::1]:7655/ws",
            "ws://[fe80::1]:7655/ws"
        )

        for (ws in localUrls) {
            val result = PairingValidator.validate(PairingPayload(v = 1, ws = ws, token = "tok-123"))
            assertTrue("Expected $ws to be valid, got $result", result is PairingValidationResult.Valid)
            val valid = result as PairingValidationResult.Valid
            assertTrue("Expected $ws to be local", valid.isLocal)
            assertFalse("Expected $ws to be cleartext ws", valid.isSecure)
        }
    }

    @Test
    fun testRejectsUnsupportedVersionEmptyTokenBadSchemesAndCleartextRemote() {
        assertEquals(
            PairingValidationResult.Invalid(PairingValidationError.UNSUPPORTED_VERSION),
            PairingValidator.validate(PairingPayload(v = 2, ws = "ws://192.168.1.1:7655/ws", token = "tok"))
        )

        assertEquals(
            PairingValidationResult.Invalid(PairingValidationError.EMPTY_TOKEN),
            PairingValidator.validate(PairingPayload(v = 1, ws = "ws://192.168.1.1:7655/ws", token = "   "))
        )

        assertEquals(
            PairingValidationResult.Invalid(PairingValidationError.UNSUPPORTED_SCHEME),
            PairingValidator.validate(PairingPayload(v = 1, ws = "http://192.168.1.1:7655/ws", token = "tok"))
        )

        assertEquals(
            PairingValidationResult.Invalid(PairingValidationError.UNSUPPORTED_SCHEME),
            PairingValidator.validate(PairingPayload(v = 1, ws = "https://example.com/ws", token = "tok"))
        )

        assertEquals(
            PairingValidationResult.Invalid(PairingValidationError.INVALID_URL),
            PairingValidator.validate(PairingPayload(v = 1, ws = "ws://user:pass@192.168.1.1:7655/ws", token = "tok"))
        )

        val remoteCleartext = listOf(
            "ws://8.8.8.8:7655/ws",
            "ws://evil.example.com/ws",
            "ws://172.32.0.1:7655/ws",
            "ws://100.128.0.1:7655/ws",
            "ws://[2001:4860:4860::8888]:7655/ws"
        )
        for (ws in remoteCleartext) {
            assertEquals(
                "Expected $ws to be rejected as cleartext remote",
                PairingValidationResult.Invalid(PairingValidationError.CLEARTEXT_REMOTE_NOT_ALLOWED),
                PairingValidator.validate(PairingPayload(v = 1, ws = ws, token = "tok"))
            )
        }
    }

    @Test
    fun testRemoteWssAllowedButMarkedNonLocal() {
        val result = PairingValidator.validate(
            PairingPayload(v = 1, ws = "wss://tunnel.example.com:443/ws", token = "tok")
        )
        assertTrue(result is PairingValidationResult.Valid)
        val valid = result as PairingValidationResult.Valid
        assertFalse(valid.isLocal)
        assertTrue(valid.isSecure)
        assertEquals("tunnel.example.com", valid.host)
    }

    @Test
    fun testIpv4MappedIpv6AndInetAddressCheck() {
        assertTrue(PairingValidator.isLocalHost("::ffff:192.168.1.10"))
        assertFalse(PairingValidator.isLocalHost("::ffff:8.8.8.8"))
        assertTrue(PairingValidator.isLocalAddress(InetAddress.getByAddress(byteArrayOf(127, 0, 0, 1))))
        assertTrue(PairingValidator.isLocalAddress(InetAddress.getByAddress(byteArrayOf(10, 1, 2, 3))))
        assertFalse(PairingValidator.isLocalAddress(InetAddress.getByAddress(byteArrayOf(8, 8, 8, 8))))
    }

    @Test
    fun testOkHttpClientBlocksCleartextToNonLocalHost() {
        val client = AirvoiceApplication.createOkHttpClient()
        val request = Request.Builder().url("http://evil.example.com/ws").build()
        try {
            client.newCall(request).execute()
            fail("Expected cleartext request to non-local host to fail")
        } catch (e: IOException) {
            assertTrue(e.message?.contains("non-local") == true)
        }
    }
}
