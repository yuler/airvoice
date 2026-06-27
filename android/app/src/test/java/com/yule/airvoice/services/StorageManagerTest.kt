package com.yule.airvoice.services

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class StorageManagerTest {
    @Test
    fun testStorageMock() {
        // Simply verify storage contract mock
        var testUrl: String? = null
        var testToken: String? = null
        
        fun save(url: String, token: String) {
            testUrl = url
            testToken = token
        }

        save("ws://test", "tok")
        assertEquals("ws://test", testUrl)
        assertEquals("tok", testToken)
    }
}
