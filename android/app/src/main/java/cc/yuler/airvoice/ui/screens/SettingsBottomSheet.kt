package cc.yuler.airvoice.ui.screens

import android.os.Build
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ContentCopy
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
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
                            viewModel.showToast(strings.toastVersionCopied, isError = false)
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
                        viewModel.disconnectAndClear()
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
                        viewModel.resetSettings()
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
private fun SettingsSection(
    title: String,
    content: @Composable ColumnScope.() -> Unit
) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(
            text = title,
            fontSize = 12.sp,
            fontWeight = FontWeight.Medium,
            color = secondaryTextColor(),
            modifier = Modifier.padding(horizontal = 4.dp)
        )
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(12.dp))
                .background(secondaryBackgroundColor())
                .border(1.dp, borderColor(), RoundedCornerShape(12.dp)),
            content = content
        )
    }
}

@Composable
private fun SettingsDivider() {
    HorizontalDivider(
        color = borderColor(),
        thickness = 1.dp,
        modifier = Modifier.padding(horizontal = 14.dp)
    )
}

@Composable
private fun SettingsToggleRow(
    label: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onCheckedChange(!checked) }
            .padding(horizontal = 14.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(
            text = label,
            fontSize = 14.sp,
            color = primaryTextColor(),
            modifier = Modifier.weight(1f)
        )
        Switch(
            checked = checked,
            onCheckedChange = onCheckedChange,
            colors = SwitchDefaults.colors(
                checkedThumbColor = Color.White,
                checkedTrackColor = AppColors.accent,
                uncheckedThumbColor = secondaryTextColor(),
                uncheckedTrackColor = chipBackgroundColor(),
                uncheckedBorderColor = borderColor()
            )
        )
    }
}

@Composable
private fun SettingsLabelledControl(
    label: String,
    control: @Composable () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 14.dp, vertical = 12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        Text(
            text = label,
            fontSize = 14.sp,
            color = primaryTextColor()
        )
        control()
    }
}

@Composable
private fun <T> SegmentedSelector(
    options: List<Pair<T, String>>,
    selected: T,
    onSelect: (T) -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(10.dp))
            .background(chipBackgroundColor())
            .padding(3.dp),
        horizontalArrangement = Arrangement.spacedBy(4.dp)
    ) {
        options.forEach { (value, label) ->
            val isSelected = value == selected
            Box(
                modifier = Modifier
                    .weight(1f)
                    .clip(RoundedCornerShape(8.dp))
                    .background(if (isSelected) AppColors.accent else Color.Transparent)
                    .clickable { onSelect(value) }
                    .padding(vertical = 7.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = label,
                    fontSize = 13.sp,
                    fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
                    color = if (isSelected) Color.White else primaryTextColor(),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
            }
        }
    }
}

@Composable
private fun SettingsActionRow(
    label: String,
    textColor: Color,
    enabled: Boolean,
    onClick: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(enabled = enabled, onClick = onClick)
            .padding(horizontal = 14.dp, vertical = 13.dp)
    ) {
        Text(
            text = label,
            fontSize = 14.sp,
            fontWeight = FontWeight.Medium,
            color = if (enabled) textColor else textColor.copy(alpha = 0.35f)
        )
    }
}

private fun formatSeconds(ms: Long): String {
    val tenths = ms / 100
    return if (tenths % 10 == 0L) "${tenths / 10}s" else "${tenths / 10}.${tenths % 10}s"
}
