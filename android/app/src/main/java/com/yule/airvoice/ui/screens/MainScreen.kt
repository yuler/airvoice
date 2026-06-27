package com.yule.airvoice.ui.screens

import android.widget.Toast
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.platform.LocalContext
import com.yule.airvoice.ui.viewmodel.AirvoiceViewModel
import com.yule.airvoice.ui.viewmodel.Screen

@Composable
fun MainScreen(viewModel: AirvoiceViewModel) {
    val context = LocalContext.current
    val currentScreen by viewModel.currentScreen.collectAsState()

    LaunchedEffect(key1 = true) {
        viewModel.toastEvents.collect { message ->
            Toast.makeText(context, message, Toast.LENGTH_SHORT).show()
        }
    }

    when (currentScreen) {
        Screen.ONBOARDING -> OnboardingScreen(
            onStartScanning = { viewModel.navigateTo(Screen.SCANNER) }
        )
        Screen.SCANNER -> QRScannerScreen(
            onQrCodeScanned = { payload -> viewModel.pairAndConnect(payload) },
            onCancel = { viewModel.navigateTo(Screen.ONBOARDING) }
        )
        Screen.HOME -> HomeScreen(viewModel = viewModel)
    }
}
