package com.yule.airvoice

import android.app.Application
import okhttp3.OkHttpClient

class AirvoiceApplication : Application() {
    val okHttpClient: OkHttpClient by lazy {
        OkHttpClient()
    }
}
