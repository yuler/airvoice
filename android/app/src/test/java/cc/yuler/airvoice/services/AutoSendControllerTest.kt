package cc.yuler.airvoice.services

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.runBlocking
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import java.lang.reflect.Field

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

    @Test
    fun testCountdownStopsWhileInFlight() = runBlocking {
        val controller = AutoSendController(textFlow, connectionManager) { _, _, _ -> }

        controller.textDidChange("hello")
        assertTrue(controller.countdownActive.value)

        setInFlight(controller, true)
        controller.textDidChange("hello world")

        assertFalse(controller.countdownActive.value)
        assertEquals(1, controller.countdownToken.value)
    }

    private fun setInFlight(controller: AutoSendController, value: Boolean) {
        val field: Field = AutoSendController::class.java.getDeclaredField("_inFlight")
        field.isAccessible = true
        @Suppress("UNCHECKED_CAST")
        val flow = field.get(controller) as MutableStateFlow<Boolean>
        flow.value = value
    }
}
