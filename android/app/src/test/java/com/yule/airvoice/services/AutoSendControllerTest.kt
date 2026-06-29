package com.yule.airvoice.services

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.runBlocking
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@RunWith(RobolectricTestRunner::class)
class AutoSendControllerTest {
    private lateinit var connectionManager: ConnectionManager
    private lateinit var textFlow: MutableStateFlow<String>

    @Before
    fun setUp() {
        connectionManager = ConnectionManager(okhttp3.OkHttpClient())
        textFlow = MutableStateFlow("")
    }

    @Test
    fun testCountdownStateTransitions() = runBlocking {
        var ackSuccess = false
        var ackTrigger: SendTrigger? = null
        val controller = AutoSendController(textFlow, connectionManager) { success, _, trigger ->
            ackSuccess = success
            ackTrigger = trigger
        }

        assertFalse(controller.countdownActive.value)
        assertEquals(0, controller.countdownToken.value)

        controller.textDidChange("hello")
        assertTrue(controller.countdownActive.value)
        assertEquals(1, controller.countdownToken.value)

        controller.textDidChange("hello world")
        assertTrue(controller.countdownActive.value)
        assertEquals(2, controller.countdownToken.value)
    }
}
