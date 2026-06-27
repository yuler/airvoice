package com.yule.airvoice.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

object AppColors {
    val lightBackground = Color(0xFFF5F5F7)
    val darkBackground = Color(0xFF000000)
    val lightSecondaryBackground = Color(0xFFFFFFFF)
    val darkSecondaryBackground = Color(0xFF0d0e15)
    val lightBorder = Color(0xFFE5E5EA)
    val darkBorder = Color(0xFF2e2e2e)
    val lightPrimaryText = Color(0xFF1C1C1E)
    val darkPrimaryText = Color(0xFFFFFFFF)
    val lightSecondaryText = Color(0xFF6C6C70)
    val darkSecondaryText = Color(0xFF8E8E93)
    val accent = Color(0xFF006efe)
    val lightChipBackground = Color(0x1A000000)
    val darkChipBackground = Color(0x1AFFFFFF)
    val lightSendButtonBackground = Color(0x0D000000)
    val darkSendButtonBackground = Color(0x1FFFFFFF)
    val lightPlaceholderText = Color(0xFFAEAEB2)
    val darkPlaceholderText = Color(0xFF8E8E93)
    val statusBarConnected = Color(0xFF00ac3a)
    val statusBarConnecting = Color(0xFFFFAE00)
    val statusBarError = Color(0xFFe2162a)
    val statusBarDisconnected = Color(0xFF8f8f8f)
}

@Composable
fun isDarkTheme(): Boolean = isSystemInDarkTheme()

@Composable
fun backgroundColor(): Color =
    if (isDarkTheme()) AppColors.darkBackground else AppColors.lightBackground

@Composable
fun secondaryBackgroundColor(): Color =
    if (isDarkTheme()) AppColors.darkSecondaryBackground else AppColors.lightSecondaryBackground

@Composable
fun borderColor(): Color =
    if (isDarkTheme()) AppColors.darkBorder else AppColors.lightBorder

@Composable
fun primaryTextColor(): Color =
    if (isDarkTheme()) AppColors.darkPrimaryText else AppColors.lightPrimaryText

@Composable
fun secondaryTextColor(): Color =
    if (isDarkTheme()) AppColors.darkSecondaryText else AppColors.lightSecondaryText

@Composable
fun chipBackgroundColor(): Color =
    if (isDarkTheme()) AppColors.darkChipBackground else AppColors.lightChipBackground

@Composable
fun sendButtonBackgroundColor(): Color =
    if (isDarkTheme()) AppColors.darkSendButtonBackground else AppColors.lightSendButtonBackground

@Composable
fun placeholderTextColor(): Color =
    if (isDarkTheme()) AppColors.darkPlaceholderText else AppColors.lightPlaceholderText
