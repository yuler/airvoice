package com.yule.airvoice.ui.theme

import androidx.compose.runtime.Composable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.Color

val LocalIsDarkTheme = staticCompositionLocalOf { false }

object AppColors {
    val lightBackground = Color(0xFFF5F5F7)
    val darkBackground = Color(0xFF000000)
    val lightSecondaryBackground = Color(0xFFFFFFFF)
    val darkSecondaryBackground = Color(0xFF0D0E15)
    val lightBorder = Color(0xFFE5E5EA)
    val darkBorder = Color(0xFF2E2E2E)
    val lightPrimaryText = Color(0xFF1C1C1E)
    val darkPrimaryText = Color(0xFFFFFFFF)
    val lightSecondaryText = Color(0xFF6C6C70)
    val darkSecondaryText = Color(0xFF8E8E93)
    val accent = Color(0xFF006EFE)
    val lightChipBackground = Color(0x0F000000) // Match iOS 6%
    val darkChipBackground = Color(0x1AFFFFFF)  // Match iOS 10%
    val lightSendButtonBackground = Color(0x0D000000) // Match iOS 5%
    val darkSendButtonBackground = Color(0x1FFFFFFF)  // Match iOS 12%
    val lightPlaceholderText = Color(0xFFAEAEB2)
    val darkPlaceholderText = Color(0xFF8E8E93)
    val statusBarConnected = Color(0xFF00AC3A)
    val statusBarConnecting = Color(0xFFFFAE00)
    val statusBarError = Color(0xFFE2162A)
    val statusBarDisconnected = Color(0xFF8F8F8F)
    
    // Toast colors
    val lightToastBackground = Color(0xEB1C1C1E) // Match iOS 92%
    val darkToastBackground = Color(0xF21F2030)  // Match iOS 95%
    
    // Tips background
    val lightTipsBackground = Color(0xFFEFEFF4)
    val darkTipsBackground = Color(0x0FFFFFFF) // Match iOS 6%
    
    // Countdown bar background
    val lightCountdownBar = Color(0x66FFAE00) // Match iOS 40%
    val darkCountdownBar = Color(0x73FFAE00)  // Match iOS 45%
}

@Composable
fun isDarkTheme(): Boolean = LocalIsDarkTheme.current

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

@Composable
fun toastBackgroundColor(): Color =
    if (isDarkTheme()) AppColors.darkToastBackground else AppColors.lightToastBackground

@Composable
fun tipsBackgroundColor(): Color =
    if (isDarkTheme()) AppColors.darkTipsBackground else AppColors.lightTipsBackground

@Composable
fun countdownBarColor(): Color =
    if (isDarkTheme()) AppColors.darkCountdownBar else AppColors.lightCountdownBar
