package cc.yuler.airvoice.models

import kotlinx.serialization.Serializable

@Serializable
data class ProtocolMessage(
    val type: String,
    val id: String? = null,
    val device: String? = null,
    val app: String? = null,
    val content: String? = null,
    val ts: Long? = null,
    val host: String? = null,
    val version: String? = null,
    val ok: Boolean? = null,
    val message: String? = null
)
