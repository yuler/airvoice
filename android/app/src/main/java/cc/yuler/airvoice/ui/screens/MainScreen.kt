package cc.yuler.airvoice.ui.screens

import android.app.Activity
import android.view.WindowManager
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cc.yuler.airvoice.models.PairingValidationResult
import cc.yuler.airvoice.models.ThemeMode
import cc.yuler.airvoice.services.ConnectionStatus
import cc.yuler.airvoice.ui.i18n.LocalAppStrings
import cc.yuler.airvoice.ui.i18n.stringsFor
import cc.yuler.airvoice.ui.theme.AppColors
import cc.yuler.airvoice.ui.theme.LocalIsDarkTheme
import cc.yuler.airvoice.ui.theme.primaryTextColor
import cc.yuler.airvoice.ui.theme.secondaryBackgroundColor
import cc.yuler.airvoice.ui.theme.secondaryTextColor
import cc.yuler.airvoice.ui.theme.toastBackgroundColor
import cc.yuler.airvoice.ui.viewmodel.AirvoiceViewModel
import cc.yuler.airvoice.ui.viewmodel.Screen

@Composable
fun MainScreen(viewModel: AirvoiceViewModel) {

    val settings by viewModel.settings.collectAsState()
    val status by viewModel.connectionManager.status.collectAsState()
    val toastMessage by viewModel.toastMessage.collectAsState()
    val isToastError by viewModel.isToastError.collectAsState()

    var currentScreen by rememberSaveable { mutableStateOf(Screen.HOME) }
    var showSettings by remember { mutableStateOf(false) }
    var pendingRemotePairing by remember { mutableStateOf<PairingValidationResult.Valid?>(null) }

    val systemDark = isSystemInDarkTheme()
    val isDark = when (settings.themeMode) {
        ThemeMode.SYSTEM -> systemDark
        ThemeMode.LIGHT -> false
        ThemeMode.DARK -> true
    }
    val configuration = LocalConfiguration.current
    val appStrings = remember(settings.language, configuration) { stringsFor(settings.language) }

    val context = LocalContext.current
    val activity = context as? Activity
    val keepAwake = (status is ConnectionStatus.Connected) && settings.keepScreenAwake
    DisposableEffect(activity, keepAwake) {
        val window = activity?.window
        if (keepAwake) {
            window?.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
        } else {
            window?.clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
        }
        onDispose {
            window?.clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
        }
    }

    CompositionLocalProvider(
        LocalIsDarkTheme provides isDark,
        LocalAppStrings provides appStrings
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            when (currentScreen) {
                Screen.SCANNER -> QRScannerScreen(
                    onQrCodeScanned = { validated ->
                        if (validated.isLocal) {
                            viewModel.pairAndConnect(validated.payload)
                        } else {
                            pendingRemotePairing = validated
                        }
                        currentScreen = Screen.HOME
                    },
                    onScanError = { message ->
                        viewModel.showToast(message, isError = true)
                    },
                    onCancel = {
                        currentScreen = Screen.HOME
                    }
                )
                Screen.HOME -> HomeScreen(
                    viewModel = viewModel,
                    isSettingsOpen = showSettings,
                    onScanQr = {
                        currentScreen = Screen.SCANNER
                    },
                    onOpenSettings = {
                        showSettings = true
                    }
                )
            }

            if (showSettings) {
                SettingsBottomSheet(
                    viewModel = viewModel,
                    onDismiss = { showSettings = false }
                )
            }

            pendingRemotePairing?.let { pending ->
                AlertDialog(
                    onDismissRequest = { pendingRemotePairing = null },
                    containerColor = secondaryBackgroundColor(),
                    titleContentColor = primaryTextColor(),
                    textContentColor = secondaryTextColor(),
                    title = { Text(appStrings.remoteAlertTitle) },
                    text = { Text(appStrings.remoteAlertBody(pending.host)) },
                    confirmButton = {
                        TextButton(
                            onClick = {
                                viewModel.pairAndConnect(pending.payload, allowRemote = true)
                                pendingRemotePairing = null
                            }
                        ) {
                            Text(appStrings.remoteAlertConfirm, color = AppColors.accent)
                        }
                    },
                    dismissButton = {
                        TextButton(onClick = { pendingRemotePairing = null }) {
                            Text(appStrings.cancel, color = secondaryTextColor())
                        }
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
