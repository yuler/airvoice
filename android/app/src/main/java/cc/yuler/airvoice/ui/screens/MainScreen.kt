package cc.yuler.airvoice.ui.screens

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cc.yuler.airvoice.ui.theme.LocalIsDarkTheme
import cc.yuler.airvoice.ui.theme.toastBackgroundColor
import cc.yuler.airvoice.ui.viewmodel.AirvoiceViewModel
import cc.yuler.airvoice.ui.viewmodel.Screen

@Composable
fun MainScreen(viewModel: AirvoiceViewModel) {
    val hasSeenOnboarding by viewModel.hasSeenOnboarding.collectAsState()
    val appTheme by viewModel.appTheme.collectAsState()
    val toastMessage by viewModel.toastMessage.collectAsState()
    val isToastError by viewModel.isToastError.collectAsState()

    var currentScreen by remember { mutableStateOf(Screen.HOME) }

    LaunchedEffect(hasSeenOnboarding) {
        currentScreen = if (hasSeenOnboarding) Screen.HOME else Screen.ONBOARDING
    }

    val isDark = appTheme == "dark"

    CompositionLocalProvider(LocalIsDarkTheme provides isDark) {
        Box(modifier = Modifier.fillMaxSize()) {
            when (currentScreen) {
                Screen.ONBOARDING -> OnboardingScreen(
                    onStartScanning = {
                        viewModel.completeOnboarding()
                    },
                    onToggleTheme = {
                        viewModel.toggleTheme()
                    }
                )
                Screen.SCANNER -> QRScannerScreen(
                    onQrCodeScanned = { payload ->
                        viewModel.pairAndConnect(payload)
                        currentScreen = Screen.HOME
                    },
                    onCancel = {
                        currentScreen = Screen.HOME
                    }
                )
                Screen.HOME -> HomeScreen(
                    viewModel = viewModel,
                    onScanQr = {
                        currentScreen = Screen.SCANNER
                    }
                )
            }

            // Custom Toast Overlay (Parity with iOS Utilities/Toast.swift)
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(bottom = 50.dp),
                contentAlignment = Alignment.BottomCenter
            ) {
                AnimatedVisibility(
                    visible = toastMessage != null,
                    enter = slideInVertically(initialOffsetY = { it / 2 }) + fadeIn(),
                    exit = slideOutVertically(targetOffsetY = { it / 2 }) + fadeOut()
                ) {
                    toastMessage?.let { msg ->
                        ToastView(message = msg, isError = isToastError)
                    }
                }
            }
        }
    }
}

@Composable
fun ToastView(message: String, isError: Boolean) {
    Row(
        modifier = Modifier
            .wrapContentSize()
            .clip(RoundedCornerShape(22.dp))
            .background(toastBackgroundColor())
            .border(1.dp, Color.White.copy(alpha = 0.1f), RoundedCornerShape(22.dp))
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        Icon(
            imageVector = if (isError) Icons.Default.Warning else Icons.Default.CheckCircle,
            contentDescription = null,
            tint = if (isError) Color.Red else Color.Green,
            modifier = Modifier.size(16.dp)
        )

        Text(
            text = message,
            fontSize = 14.sp,
            color = Color.White,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis
        )
    }
}
