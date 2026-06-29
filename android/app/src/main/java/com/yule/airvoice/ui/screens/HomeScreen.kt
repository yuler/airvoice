package com.yule.airvoice.ui.screens

import android.content.Intent
import android.net.Uri
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.LinearEasing
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
import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.material.icons.filled.Info
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.yule.airvoice.services.ConnectionStatus
import com.yule.airvoice.ui.theme.*
import com.yule.airvoice.ui.viewmodel.AirvoiceViewModel

@Composable
fun HomeScreen(
    viewModel: AirvoiceViewModel,
    onScanQr: () -> Unit = {}
) {
    val text by viewModel.inputText.collectAsState()
    val status by viewModel.connectionManager.status.collectAsState()
    val countdownActive by viewModel.autoSendController?.countdownActive?.collectAsState() ?: remember { mutableStateOf(false) }
    val countdownToken by viewModel.autoSendController?.countdownToken?.collectAsState() ?: remember { mutableStateOf(0) }
    val inFlight by viewModel.autoSendController?.inFlight?.collectAsState() ?: remember { mutableStateOf(false) }
    val sendTimedOut by viewModel.sendTimedOut.collectAsState()
    val appTheme by viewModel.appTheme.collectAsState()

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
    val shouldBlinkStatusDot = isConnected && !sendTimedOut

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
                    is ConnectionStatus.Connected -> if (!sendTimedOut) AppColors.statusBarConnected else AppColors.statusBarConnecting
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

                if (shouldBlinkStatusDot) {
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

                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    // Theme toggle button
                    IconButton(
                        onClick = { viewModel.toggleTheme() },
                        modifier = Modifier
                            .size(28.dp)
                            .background(chipBackgroundColor(), CircleShape)
                    ) {
                        Text(
                            text = if (appTheme == "light") "🌙" else "☀️",
                            fontSize = 13.sp
                        )
                    }

                    // QR scanner button
                    IconButton(
                        onClick = onScanQr,
                        modifier = Modifier
                            .size(28.dp)
                            .background(chipBackgroundColor(), CircleShape)
                    ) {
                        Text(
                            text = "📷",
                            fontSize = 13.sp
                        )
                    }
                }
            }

            // AutoSendCountdownBar (drains from 1 to 0 over 1.5s)
            AutoSendCountdownBar(
                active = countdownActive,
                token = countdownToken,
                duration = 1500L
            )

            // Main content
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .padding(top = 12.dp),
                verticalArrangement = Arrangement.spacedBy(20.dp)
            ) {
                // Editor
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .weight(1f)
                        .padding(horizontal = 20.dp)
                        .border(1.dp, borderClr, RoundedCornerShape(16.dp))
                        .background(editorBg, RoundedCornerShape(16.dp))
                ) {
                    if (text.isEmpty()) {
                        Text(
                            text = "在此输入，或使用键盘麦克风语音输入...",
                            color = placeholderClr,
                            modifier = Modifier
                                .padding(horizontal = 16.dp, vertical = 16.dp)
                        )
                    }
                    TextField(
                        value = text,
                        onValueChange = { viewModel.updateInputText(it) },
                        modifier = Modifier
                            .fillMaxSize()
                            .focusRequester(focusRequester),
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
                        .padding(bottom = 20.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    if (inFlight) {
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(6.dp),
                            modifier = Modifier
                                .background(chipBackgroundColor(), RoundedCornerShape(14.dp))
                                .padding(horizontal = 12.dp, vertical = 6.dp)
                        ) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(12.dp),
                                strokeWidth = 2.dp,
                                color = subTextColor
                            )
                            Text("发送中", fontSize = 12.sp, color = subTextColor)
                            TextButton(
                                onClick = { viewModel.cancelSend() },
                                contentPadding = PaddingValues(0.dp),
                                modifier = Modifier.height(20.dp)
                            ) {
                                Text("取消", fontSize = 12.sp, fontWeight = FontWeight.SemiBold, color = AppColors.accent)
                            }
                        }
                    } else if (!isConnected) {
                        Text(
                            text = "请先扫码连接电脑",
                            fontSize = 12.sp,
                            color = subTextColor
                        )
                    }

                      Button(
                          onClick = {
                              viewModel.manualSend()
                          },
                          modifier = Modifier
                              .fillMaxWidth()
                              .height(44.dp),
                          shape = RoundedCornerShape(22.dp),
                          colors = ButtonDefaults.buttonColors(
                              containerColor = sendBtnBg,
                              disabledContainerColor = sendBtnBg.copy(alpha = 0.5f)
                          ),
                          enabled = isConnected && !inFlight
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

                      InputMethodTipsView()
                  }
              }
          }
      }
  }

  @Composable
  fun AutoSendCountdownBar(active: Boolean, token: Int, duration: Long) {
      val progress = remember { Animatable(0f) }

      LaunchedEffect(token) {
          if (active) {
              progress.snapTo(1f)
              progress.animateTo(
                  targetValue = 0f,
                  animationSpec = tween(durationMillis = duration.toInt(), easing = LinearEasing)
              )
          } else {
              progress.snapTo(0f)
          }
      }

      Box(
          modifier = Modifier
              .fillMaxWidth()
              .height(3.dp)
              .background(Color.Transparent)
      ) {
          if (active) {
              Box(
                  modifier = Modifier
                      .fillMaxHeight()
                      .fillMaxWidth(progress.value)
                      .background(countdownBarColor())
              )
          }
      }
  }

  @Composable
  fun InputMethodTipsView() {
      val tipsBg = tipsBackgroundColor()
      val textColor = primaryTextColor()
      val subTextColor = secondaryTextColor()

      Column(
          modifier = Modifier
              .fillMaxWidth()
              .clip(RoundedCornerShape(12.dp))
              .background(tipsBg)
              .padding(14.dp),
          verticalArrangement = Arrangement.spacedBy(10.dp)
      ) {
          Row(
              verticalAlignment = Alignment.CenterVertically,
              horizontalArrangement = Arrangement.spacedBy(8.dp)
          ) {
              Icon(
                  imageVector = Icons.Default.Info,
                  contentDescription = null,
                  modifier = Modifier.size(16.dp),
                  tint = textColor
              )
              Text(
                  "语音输入需要第三方输入法",
                  fontSize = 12.sp,
                  fontWeight = FontWeight.SemiBold,
                  color = textColor
              )
          }

          Text(
              "优先使用豆包或微信输入法，在键盘中点击麦克风说话：",
              fontSize = 12.sp,
              color = subTextColor
          )

          Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
              AppMarketLinkButton("豆包输入法", "https://www.doubao.com/")
              AppMarketLinkButton("微信输入法", "https://z.weixin.qq.com/")
          }

          Text(
              "安装后前往 系统设置 → 语言与输入法 启用并切换键盘",
              fontSize = 10.sp,
              color = subTextColor
          )
      }
  }

  @Composable
  fun AppMarketLinkButton(title: String, url: String) {
      val context = LocalContext.current
      val accent = AppColors.accent
      val isDark = LocalIsDarkTheme.current
      val btnBg = accent.copy(alpha = if (isDark) 0.15f else 0.10f)

      TextButton(
          onClick = {
              val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
              context.startActivity(intent)
          },
          contentPadding = PaddingValues(horizontal = 10.dp, vertical = 6.dp),
          colors = ButtonDefaults.textButtonColors(
              containerColor = btnBg
          ),
          modifier = Modifier.height(28.dp),
          shape = RoundedCornerShape(8.dp)
      ) {
          Text(
              text = "↗ $title",
              fontSize = 12.sp,
              fontWeight = FontWeight.Medium,
              color = accent
          )
      }
  }
