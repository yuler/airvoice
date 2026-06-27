package com.yule.airvoice.ui.screens

import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.yule.airvoice.services.ConnectionStatus
import com.yule.airvoice.ui.theme.AppColors
import com.yule.airvoice.ui.theme.backgroundColor
import com.yule.airvoice.ui.theme.borderColor
import com.yule.airvoice.ui.theme.chipBackgroundColor
import com.yule.airvoice.ui.theme.placeholderTextColor
import com.yule.airvoice.ui.theme.primaryTextColor
import com.yule.airvoice.ui.theme.secondaryBackgroundColor
import com.yule.airvoice.ui.theme.secondaryTextColor
import com.yule.airvoice.ui.theme.sendButtonBackgroundColor
import com.yule.airvoice.ui.viewmodel.AirvoiceViewModel

@Composable
fun HomeScreen(
    viewModel: AirvoiceViewModel,
    onScanQr: () -> Unit = {}
) {
    val text by viewModel.inputText.collectAsState()
    val status by viewModel.connectionManager.status.collectAsState()
    val focusRequester = remember { FocusRequester() }

    val bgColor = backgroundColor()
    val textColor = primaryTextColor()
    val subTextColor = secondaryTextColor()
    val editorBg = secondaryBackgroundColor()
    val borderClr = borderColor()
    val placeholderClr = placeholderTextColor()
    val sendBtnBg = sendButtonBackgroundColor()

    val density = LocalDensity.current
    val imeInsets = WindowInsets.ime
    LaunchedEffect(Unit) {
        kotlinx.coroutines.delay(350)
        focusRequester.requestFocus()
    }
    LaunchedEffect(Unit) {
        var prevImeVisible = false
        snapshotFlow { imeInsets.getBottom(density) > 0 }
            .collect { isImeVisible ->
                if (prevImeVisible && !isImeVisible) {
                    viewModel.triggerImmediateSend()
                }
                prevImeVisible = isImeVisible
            }
    }

    val isConnected = status is ConnectionStatus.Connected

    val infiniteTransition = rememberInfiniteTransition(label = "breathing")
    val breathAlpha by infiniteTransition.animateFloat(
        initialValue = 1f,
        targetValue = 0.4f,
        animationSpec = infiniteRepeatable(
            animation = tween(1000),
            repeatMode = RepeatMode.Reverse
        ),
        label = "breathAlpha"
    )
    val breathScale by infiniteTransition.animateFloat(
        initialValue = 1f,
        targetValue = 1.25f,
        animationSpec = infiniteRepeatable(
            animation = tween(1000),
            repeatMode = RepeatMode.Reverse
        ),
        label = "breathScale"
    )

    Box(modifier = Modifier.fillMaxSize().background(bgColor)) {
        Column(modifier = Modifier.fillMaxSize()) {
            // Status bar
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                val dotColor = when (status) {
                    is ConnectionStatus.Connected -> AppColors.statusBarConnected
                    is ConnectionStatus.Connecting -> AppColors.statusBarConnecting
                    is ConnectionStatus.Error -> AppColors.statusBarError
                    ConnectionStatus.Disconnected -> AppColors.statusBarDisconnected
                }
                val label = when (status) {
                    is ConnectionStatus.Connected -> "已连接: ${(status as ConnectionStatus.Connected).host}"
                    is ConnectionStatus.Connecting -> "连接中..."
                    is ConnectionStatus.Error -> "连接失败，正在重试"
                    ConnectionStatus.Disconnected -> "未连接"
                }

                if (isConnected) {
                    Box(
                        modifier = Modifier
                            .size(8.dp)
                            .scale(breathScale)
                            .clip(CircleShape)
                            .background(dotColor.copy(alpha = breathAlpha))
                    )
                } else {
                    Box(
                        modifier = Modifier
                            .size(8.dp)
                            .clip(CircleShape)
                            .background(dotColor)
                    )
                }
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = label,
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Medium,
                    color = textColor.copy(alpha = 0.8f)
                )

                Spacer(modifier = Modifier.weight(1f))

                // QR scanner button
                IconButton(
                    onClick = onScanQr,
                    modifier = Modifier
                        .size(28.dp)
                        .background(chipBackgroundColor(), CircleShape)
                ) {
                    Icon(
                        imageVector = Icons.Default.Search,
                        contentDescription = "扫码",
                        tint = textColor,
                        modifier = Modifier.size(14.dp)
                    )
                }
                Spacer(modifier = Modifier.width(8.dp))
                // Re-pair button
                IconButton(
                    onClick = { viewModel.disconnectAndClear() },
                    modifier = Modifier
                        .size(28.dp)
                        .background(chipBackgroundColor(), CircleShape)
                ) {
                    Icon(
                        imageVector = Icons.Default.Refresh,
                        contentDescription = "重新配对",
                        tint = textColor,
                        modifier = Modifier.size(14.dp)
                    )
                }
            }

            // Editor
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .padding(horizontal = 20.dp)
                    .border(1.dp, borderClr, RoundedCornerShape(16.dp))
                    .background(editorBg, RoundedCornerShape(16.dp))
            ) {
                TextField(
                    value = text,
                    onValueChange = { viewModel.updateInputText(it) },
                    modifier = Modifier
                        .fillMaxSize()
                        .focusRequester(focusRequester),
                    placeholder = {
                        Text(
                            text = "在此输入，或使用键盘麦克风语音输入...",
                            color = placeholderClr
                        )
                    },
                    colors = TextFieldDefaults.colors(
                        focusedContainerColor = Color.Transparent,
                        unfocusedContainerColor = Color.Transparent,
                        disabledContainerColor = Color.Transparent,
                        focusedIndicatorColor = Color.Transparent,
                        unfocusedIndicatorColor = Color.Transparent
                    ),
                    textStyle = LocalTextStyle.current.copy(color = textColor)
                )
            }

            // Bottom controls
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 20.dp)
                    .padding(bottom = 20.dp, top = 12.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                if (!isConnected) {
                    Text(
                        text = "请先扫码连接电脑",
                        fontSize = 12.sp,
                        color = subTextColor
                    )
                    Spacer(modifier = Modifier.height(12.dp))
                }

                Button(
                    onClick = {
                        viewModel.triggerImmediateSend()
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(44.dp),
                    shape = RoundedCornerShape(22.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = sendBtnBg,
                        disabledContainerColor = sendBtnBg.copy(alpha = 0.5f)
                    ),
                    enabled = isConnected
                ) {
                    Icon(
                        imageVector = Icons.AutoMirrored.Filled.Send,
                        contentDescription = null,
                        modifier = Modifier.size(16.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        "发送到电脑",
                        fontSize = 15.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }
        }
    }
}
