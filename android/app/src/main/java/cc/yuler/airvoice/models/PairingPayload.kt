package cc.yuler.airvoice.models

import kotlinx.serialization.Serializable
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import java.net.Inet4Address
import java.net.Inet6Address
import java.net.InetAddress

@Serializable
data class PairingPayload(
    val v: Int,
    val ws: String,
    val token: String
)

enum class PairingValidationError {
    UNSUPPORTED_VERSION,
    EMPTY_TOKEN,
    INVALID_URL,
    UNSUPPORTED_SCHEME,
    CLEARTEXT_REMOTE_NOT_ALLOWED
}

sealed class PairingValidationResult {
    data class Valid(
        val payload: PairingPayload,
        val normalizedUrl: String,
        val host: String,
        val isLocal: Boolean,
        val isSecure: Boolean
    ) : PairingValidationResult()

    data class Invalid(val reason: PairingValidationError) : PairingValidationResult()
}

object PairingValidator {
    fun errorMessage(error: PairingValidationError): String = when (error) {
        PairingValidationError.UNSUPPORTED_VERSION -> "不支持的二维码协议版本"
        PairingValidationError.EMPTY_TOKEN -> "二维码缺少认证令牌"
        PairingValidationError.INVALID_URL -> "无效的服务器地址"
        PairingValidationError.UNSUPPORTED_SCHEME -> "仅支持 ws:// 或 wss:// 协议"
        PairingValidationError.CLEARTEXT_REMOTE_NOT_ALLOWED -> "拒绝连接：非局域网地址不支持明文 ws://"
    }

    fun validate(payload: PairingPayload): PairingValidationResult {
        if (payload.v != 1) {
            return PairingValidationResult.Invalid(PairingValidationError.UNSUPPORTED_VERSION)
        }
        return validate(payload.ws, payload.token, payload)
    }

    fun validate(
        wsUrl: String,
        token: String,
        originalPayload: PairingPayload? = null
    ): PairingValidationResult {
        val trimmedToken = token.trim()
        if (trimmedToken.isEmpty()) {
            return PairingValidationResult.Invalid(PairingValidationError.EMPTY_TOKEN)
        }

        val trimmedUrl = wsUrl.trim()
        val schemeEnd = trimmedUrl.indexOf("://")
        if (schemeEnd <= 0) {
            return PairingValidationResult.Invalid(PairingValidationError.INVALID_URL)
        }

        val scheme = trimmedUrl.substring(0, schemeEnd).lowercase()
        if (scheme != "ws" && scheme != "wss") {
            return PairingValidationResult.Invalid(PairingValidationError.UNSUPPORTED_SCHEME)
        }

        val isSecure = scheme == "wss"
        val httpPrefix = if (isSecure) "https://" else "http://"
        val httpCandidate = httpPrefix + trimmedUrl.substring(schemeEnd + 3)
        val parsed = httpCandidate.toHttpUrlOrNull()
            ?: return PairingValidationResult.Invalid(PairingValidationError.INVALID_URL)

        if (parsed.username.isNotEmpty() || parsed.password.isNotEmpty()) {
            return PairingValidationResult.Invalid(PairingValidationError.INVALID_URL)
        }

        val host = parsed.host
        if (host.isBlank()) {
            return PairingValidationResult.Invalid(PairingValidationError.INVALID_URL)
        }

        val isLocal = isLocalHost(host)
        if (!isSecure && !isLocal) {
            return PairingValidationResult.Invalid(PairingValidationError.CLEARTEXT_REMOTE_NOT_ALLOWED)
        }

        val normalizedUrl = parsed.toString()
            .replaceFirst("https://", "wss://")
            .replaceFirst("http://", "ws://")

        val payload = originalPayload ?: PairingPayload(v = 1, ws = normalizedUrl, token = trimmedToken)
        return PairingValidationResult.Valid(
            payload = payload,
            normalizedUrl = normalizedUrl,
            host = host,
            isLocal = isLocal,
            isSecure = isSecure
        )
    }

    fun isLocalHost(rawHost: String): Boolean {
        var host = rawHost.trim()
        if (host.startsWith("[") && host.endsWith("]") && host.length >= 2) {
            host = host.substring(1, host.length - 1)
        }
        val zoneIdx = host.indexOf('%')
        if (zoneIdx >= 0) {
            host = host.substring(0, zoneIdx)
        }
        host = host.lowercase().trimEnd('.')
        if (host.isEmpty()) return false

        if (host == "localhost" ||
            host.endsWith(".localhost") ||
            host.endsWith(".local") ||
            host.endsWith(".lan") ||
            host.endsWith(".home.arpa")
        ) {
            return true
        }

        parseIpv4Bytes(host)?.let { bytes ->
            return isLocalIpv4(bytes)
        }

        if (host.contains(':')) {
            parseIpv6Bytes(host)?.let { bytes ->
                return isLocalIpv6(bytes)
            }
        }

        return false
    }

    fun isLocalAddress(address: InetAddress): Boolean {
        val bytes = address.address
        return when (address) {
            is Inet4Address -> bytes.size == 4 && isLocalIpv4(bytes)
            is Inet6Address -> bytes.size == 16 && isLocalIpv6(bytes)
            else -> false
        }
    }

    private fun isLocalIpv4(bytes: ByteArray): Boolean {
        val b0 = bytes[0].toInt() and 0xFF
        val b1 = bytes[1].toInt() and 0xFF
        return b0 == 127 ||                         // 127.0.0.0/8 loopback
            b0 == 10 ||                             // 10.0.0.0/8 RFC 1918
            (b0 == 172 && b1 in 16..31) ||          // 172.16.0.0/12 RFC 1918
            (b0 == 192 && b1 == 168) ||             // 192.168.0.0/16 RFC 1918
            (b0 == 169 && b1 == 254) ||             // 169.254.0.0/16 link-local
            (b0 == 100 && b1 in 64..127)            // 100.64.0.0/10 CGNAT / Tailscale
    }

    private fun isLocalIpv6(bytes: ByteArray): Boolean {
        // ::1 loopback
        val isLoopback = (0 until 15).all { bytes[it].toInt() == 0 } && bytes[15].toInt() == 1
        if (isLoopback) return true

        val b0 = bytes[0].toInt() and 0xFF
        val b1 = bytes[1].toInt() and 0xFF

        // fc00::/7 unique local addresses (includes Tailscale fd7a:115c:a1e0::/48)
        if ((b0 and 0xFE) == 0xFC) return true

        // fe80::/10 link-local
        if (b0 == 0xFE && (b1 and 0xC0) == 0x80) return true

        // ::ffff:0:0/96 IPv4-mapped IPv6
        val isIpv4Mapped = (0 until 10).all { bytes[it].toInt() == 0 } &&
            (bytes[10].toInt() and 0xFF) == 0xFF &&
            (bytes[11].toInt() and 0xFF) == 0xFF
        if (isIpv4Mapped) {
            return isLocalIpv4(bytes.copyOfRange(12, 16))
        }

        return false
    }

    private fun parseIpv4Bytes(host: String): ByteArray? {
        val parts = host.split('.')
        if (parts.size != 4) return null
        val out = ByteArray(4)
        for (i in 0 until 4) {
            val part = parts[i]
            if (part.isEmpty() || part.length > 3 || !part.all { it in '0'..'9' }) return null
            if (part.length > 1 && part[0] == '0') return null
            val v = part.toIntOrNull() ?: return null
            if (v !in 0..255) return null
            out[i] = v.toByte()
        }
        return out
    }

    private fun parseIpv6Bytes(host: String): ByteArray? {
        // Handle embedded IPv4 suffix (e.g. ::ffff:192.168.1.1)
        val normalized = if (host.contains('.')) {
            val lastColon = host.lastIndexOf(':')
            if (lastColon < 0) return null
            val v4 = parseIpv4Bytes(host.substring(lastColon + 1)) ?: return null
            val hi = ((v4[0].toInt() and 0xFF) shl 8) or (v4[1].toInt() and 0xFF)
            val lo = ((v4[2].toInt() and 0xFF) shl 8) or (v4[3].toInt() and 0xFF)
            host.substring(0, lastColon + 1) + hi.toString(16) + ":" + lo.toString(16)
        } else {
            host
        }

        val doubleColonIdx = normalized.indexOf("::")
        if (doubleColonIdx >= 0 && normalized.indexOf("::", doubleColonIdx + 2) >= 0) {
            return null
        }

        val groups: List<Int> = if (doubleColonIdx >= 0) {
            val leftPart = normalized.substring(0, doubleColonIdx)
            val rightPart = normalized.substring(doubleColonIdx + 2)
            val left = if (leftPart.isEmpty()) emptyList() else parseHexGroups(leftPart) ?: return null
            val right = if (rightPart.isEmpty()) emptyList() else parseHexGroups(rightPart) ?: return null
            val missing = 8 - left.size - right.size
            if (missing < 1) return null
            left + List(missing) { 0 } + right
        } else {
            val parsed = parseHexGroups(normalized) ?: return null
            if (parsed.size != 8) return null
            parsed
        }

        val out = ByteArray(16)
        for (i in 0 until 8) {
            out[i * 2] = ((groups[i] ushr 8) and 0xFF).toByte()
            out[i * 2 + 1] = (groups[i] and 0xFF).toByte()
        }
        return out
    }

    private fun parseHexGroups(part: String): List<Int>? {
        val tokens = part.split(':')
        val result = ArrayList<Int>(tokens.size)
        for (tok in tokens) {
            if (tok.isEmpty() || tok.length > 4) return null
            val value = tok.toIntOrNull(16) ?: return null
            if (value !in 0..0xFFFF) return null
            result.add(value)
        }
        return result
    }
}
