package cc.yuler.airvoice.ui.screens

// Shared settings UI primitives live in SettingsComponents.kt.

import android.os.Build
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ContentCopy
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cc.yuler.airvoice.models.AppLanguage
import cc.yuler.airvoice.models.AppSettings
import cc.yuler.airvoice.models.ThemeMode
import cc.yuler.airvoice.ui.i18n.LocalAppStrings
import cc.yuler.airvoice.ui.theme.*
import cc.yuler.airvoice.ui.viewmodel.AirvoiceViewModel
import java.util.Locale
import kotlinx.coroutines.delay

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsBottomSheet(
    viewModel: AirvoiceViewModel,
    onDismiss: () -> Unit
) {
    val strings = LocalAppStrings.current
    val settings by viewModel.settings.collectAsState()
    val pairedWsUrl by viewModel.pairedWsUrl.collectAsState()
    val context = LocalContext.current
    val clipboardManager = LocalClipboardManager.current

    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    var showDisconnectConfirm by remember { mutableStateOf(false) }
    var showResetConfirm by remember { mutableStateOf(false) }
    var sheetToastMessage by remember { mutableStateOf<String?>(null) }
    var sheetToastToken by remember { mutableIntStateOf(0) }

    LaunchedEffect(sheetToastToken) {
        if (sheetToastMessage != null) {
            delay(2000)
            sheetToastMessage = null
        }
    }

    fun showLocalToast(message: String) {
        sheetToastMessage = message
        sheetToastToken += 1
    }

    val appVersion = remember(context) {
        try {
            @Suppress("DEPRECATION")
            context.packageManager.getPackageInfo(context.packageName, 0).versionName ?: "0.0.0"
        } catch (_: Exception) {
            "0.0.0"
        }
    }
    val systemVersion = remember {
        "Android ${Build.VERSION.RELEASE} (API ${Build.VERSION.SDK_INT})"
    }

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
        containerColor = backgroundColor(),
        contentColor = primaryTextColor(),
        dragHandle = {
            BottomSheetDefaults.DragHandle(color = borderColor())
        }
    ) {
        Box(modifier = Modifier.fillMaxWidth()) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .navigationBarsPadding()
                    .verticalScroll(rememberScrollState())
                    .padding(horizontal = 20.dp)
                    .padding(bottom = 28.dp),
                verticalArrangement = Arrangement.spacedBy(20.dp)
            ) {
                // Header
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                        text = strings.settingsTitle,
                        fontSize = 20.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = primaryTextColor()
                    )
                    TextButton(
                        onClick = onDismiss,
                        contentPadding = PaddingValues(horizontal = 8.dp, vertical = 4.dp)
                    ) {
                        Text(
                            text = strings.done,
                            fontSize = 15.sp,
                            fontWeight = FontWeight.Medium,
                            color = AppColors.accent
                        )
                    }
                }

                // 1. Appearance & Language
                SettingsSection(title = strings.sectionAppearance) {
                    SettingsLabelledControl(label = strings.themeLabel) {
                        SegmentedSelector(
                            options = listOf(
                                ThemeMode.SYSTEM to strings.themeSystem,
                                ThemeMode.LIGHT to strings.themeLight,
                                ThemeMode.DARK to strings.themeDark
                            ),
                            selected = settings.themeMode,
                            onSelect = { viewModel.setThemeMode(it) }
                        )
                    }
                    SettingsDivider()
                    SettingsLabelledControl(label = strings.languageLabel) {
                        SegmentedSelector(
                            options = listOf(
                                AppLanguage.SYSTEM to strings.languageSystem,
                                AppLanguage.ZH to strings.languageZh,
                                AppLanguage.EN to strings.languageEn
                            ),
                            selected = settings.language,
                            onSelect = { viewModel.setLanguage(it) }
                        )
                    }
                }

                // 2. General
                SettingsSection(title = strings.sectionGeneral) {
                    SettingsToggleRow(
                        label = strings.autoSendLabel,
                        checked = settings.autoSendEnabled,
                        onCheckedChange = { viewModel.setAutoSendEnabled(it) }
                    )

                    AnimatedVisibility(visible = settings.autoSendEnabled) {
                        Column {
                            SettingsDivider()
                            SettingsLabelledControl(label = strings.autoSendDelayLabel) {
                                SegmentedSelector(
                                    options = AppSettings.AUTO_SEND_DELAY_OPTIONS.map { ms ->
                                        ms to formatSeconds(ms)
                                    },
                                    selected = settings.autoSendDelayMs,
                                    onSelect = { viewModel.setAutoSendDelayMs(it) }
                                )
                            }
                        }
                    }

                    SettingsDivider()
                    SettingsToggleRow(
                        label = strings.hapticFeedbackLabel,
                        checked = settings.hapticFeedbackEnabled,
                        onCheckedChange = { viewModel.setHapticFeedbackEnabled(it) }
                    )
                    SettingsDivider()
                    SettingsToggleRow(
                        label = strings.keepScreenAwakeLabel,
                        checked = settings.keepScreenAwake,
                        onCheckedChange = { viewModel.setKeepScreenAwake(it) }
                    )
                    SettingsDivider()
                    SettingsToggleRow(
                        label = strings.showInputTipsLabel,
                        checked = settings.showInputTips,
                        onCheckedChange = { viewModel.setShowInputTips(it) }
                    )
                    SettingsDivider()
                    SettingsLabelledControl(label = strings.sendTimeoutLabel) {
                        SegmentedSelector(
                            options = AppSettings.SEND_TIMEOUT_OPTIONS.map { ms ->
                                ms to "${ms / 1000}s"
                            },
                            selected = settings.sendTimeoutMs,
                            onSelect = { viewModel.setSendTimeoutMs(it) }
                        )
                    }
                }

                // 3. Connection & Data
                SettingsSection(title = strings.sectionConnection) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 14.dp, vertical = 12.dp),
                        verticalArrangement = Arrangement.spacedBy(4.dp)
                    ) {
                        Text(
                            text = strings.pairedServerLabel,
                            fontSize = 14.sp,
                            color = primaryTextColor()
                        )
                        Text(
                            text = pairedWsUrl ?: strings.notPaired,
                            fontSize = 13.sp,
                            fontFamily = if (pairedWsUrl != null) FontFamily.Monospace else FontFamily.Default,
                            color = secondaryTextColor(),
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis
                        )
                    }
                    SettingsDivider()
                    val hasPairing = !pairedWsUrl.isNullOrEmpty()
                    SettingsActionRow(
                        label = strings.disconnectAndClear,
                        textColor = AppColors.statusBarError,
                        enabled = hasPairing,
                        onClick = { showDisconnectConfirm = true }
                    )
                    SettingsDivider()
                    SettingsActionRow(
                        label = strings.resetDefaults,
                        textColor = primaryTextColor(),
                        enabled = true,
                        onClick = { showResetConfirm = true }
                    )
                }

                // 4. About
                SettingsSection(title = strings.sectionAbout) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable {
                                val diagnostic = "Airvoice v$appVersion · $systemVersion"
                                clipboardManager.setText(AnnotatedString(diagnostic))
                                showLocalToast(strings.toastVersionCopied)
                            }
                            .padding(horizontal = 14.dp, vertical = 12.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(
                                text = strings.appVersionLabel,
                                fontSize = 14.sp,
                                color = primaryTextColor()
                            )
                            Text(
                                text = "v$appVersion",
                                fontSize = 13.sp,
                                fontFamily = FontFamily.Monospace,
                                color = secondaryTextColor()
                            )
                        }
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(
                                text = strings.systemVersionLabel,
                                fontSize = 14.sp,
                                color = primaryTextColor()
                            )
                            Text(
                                text = systemVersion,
                                fontSize = 13.sp,
                                fontFamily = FontFamily.Monospace,
                                color = secondaryTextColor()
                            )
                        }
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.End
                        ) {
                            Icon(
                                imageVector = Icons.Default.ContentCopy,
                                contentDescription = null,
                                tint = secondaryTextColor(),
                                modifier = Modifier.size(12.dp)
                            )
                            Spacer(modifier = Modifier.width(4.dp))
                            Text(
                                text = strings.tapToCopyHint,
                                fontSize = 11.sp,
                                color = secondaryTextColor()
                            )
                        }
                    }
                }
            }

            // Local toast overlay inside ModalBottomSheet's Dialog window
            SheetToastOverlay(
                message = sheetToastMessage,
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .navigationBarsPadding()
                    .padding(bottom = 24.dp)
            )
        }
    }

    if (showDisconnectConfirm) {
        AlertDialog(
            onDismissRequest = { showDisconnectConfirm = false },
            containerColor = secondaryBackgroundColor(),
            titleContentColor = primaryTextColor(),
            textContentColor = secondaryTextColor(),
            title = { Text(strings.disconnectConfirmTitle) },
            text = { Text(strings.disconnectConfirmBody) },
            confirmButton = {
                TextButton(
                    onClick = {
                        showDisconnectConfirm = false
                        viewModel.disconnectAndClear(showToast = false)
                        showLocalToast(strings.toastPairingCleared)
                    }
                ) {
                    Text(strings.disconnectConfirmAction, color = AppColors.statusBarError)
                }
            },
            dismissButton = {
                TextButton(onClick = { showDisconnectConfirm = false }) {
                    Text(strings.cancel, color = secondaryTextColor())
                }
            }
        )
    }

    if (showResetConfirm) {
        AlertDialog(
            onDismissRequest = { showResetConfirm = false },
            containerColor = secondaryBackgroundColor(),
            titleContentColor = primaryTextColor(),
            textContentColor = secondaryTextColor(),
            title = { Text(strings.resetConfirmTitle) },
            text = { Text(strings.resetConfirmBody) },
            confirmButton = {
                TextButton(
                    onClick = {
                        showResetConfirm = false
                        viewModel.resetSettings(showToast = false)
                        showLocalToast(viewModel.currentStrings().toastSettingsReset)
                    }
                ) {
                    Text(strings.resetConfirmAction, color = AppColors.accent)
                }
            },
            dismissButton = {
                TextButton(onClick = { showResetConfirm = false }) {
                    Text(strings.cancel, color = secondaryTextColor())
                }
            }
        )
    }
}

@Composable
private fun SheetToastOverlay(
    message: String?,
    modifier: Modifier = Modifier
) {
    Box(modifier = modifier) {
        AnimatedVisibility(
            visible = message != null,
            enter = slideInVertically(initialOffsetY = { it / 2 }) + fadeIn(),
            exit = slideOutVertically(targetOffsetY = { it / 2 }) + fadeOut()
        ) {
            message?.let { msg ->
                ToastView(message = msg, isError = false)
            }
        }
    }
}

private fun formatSeconds(ms: Long): String = String.format(Locale.US, "%.1fs", ms / 1000.0)
