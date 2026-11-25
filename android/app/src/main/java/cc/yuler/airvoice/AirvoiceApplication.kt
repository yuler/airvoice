package cc.yuler.airvoice

import android.app.Application
import cc.yuler.airvoice.models.PairingValidator
import okhttp3.Interceptor
import okhttp3.OkHttpClient
import java.io.IOException

class AirvoiceApplication : Application() {
    val okHttpClient: OkHttpClient by lazy {
        createOkHttpClient()
    }

    companion object {
        fun createOkHttpClient(): OkHttpClient {
            val hostGuard = Interceptor { chain ->
                val request = chain.request()
                if (!request.isHttps && !PairingValidator.isLocalHost(request.url.host)) {
                    throw IOException("Cleartext traffic to non-local host is not permitted")
                }
                chain.proceed(request)
            }
            val resolvedAddressGuard = Interceptor { chain ->
                val request = chain.request()
                if (!request.isHttps) {
                    val address = chain.connection()?.socket()?.inetAddress
                    if (address != null && !PairingValidator.isLocalAddress(address)) {
                        throw IOException("Cleartext traffic to non-local address is not permitted")
                    }
                }
                chain.proceed(request)
            }
            return OkHttpClient.Builder()
                .addInterceptor(hostGuard)
                .addNetworkInterceptor(resolvedAddressGuard)
                .build()
        }
    }
}
