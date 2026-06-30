package com.yule.airvoice.models

import kotlinx.serialization.Serializable

@Serializable
data class PairingPayload(
    val v: Int,
    val ws: String,
    val token: String
)
