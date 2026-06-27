package com.yule.airvoice.ui.screens

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.yule.airvoice.ui.theme.AppColors
import com.yule.airvoice.ui.theme.backgroundColor
import com.yule.airvoice.ui.theme.borderColor
import com.yule.airvoice.ui.theme.primaryTextColor
import com.yule.airvoice.ui.theme.secondaryBackgroundColor
import com.yule.airvoice.ui.theme.secondaryTextColor

@Composable
fun OnboardingScreen(onStartScanning: () -> Unit) {
    val bgColor = backgroundColor()
    val textColor = primaryTextColor()
    val subTextColor = secondaryTextColor()
    val cardBg = secondaryBackgroundColor()
    val borderClr = borderColor()

    Surface(
        modifier = Modifier.fillMaxSize(),
        color = bgColor
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.weight(1f))

            // Icon
            Box(
                modifier = Modifier
                    .size(80.dp)
                    .background(AppColors.accent.copy(alpha = 0.15f), CircleShape),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "\uD83C\uDFA4",
                    fontSize = 36.sp
                )
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Title
            Text(
                text = "Airvoice",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = textColor
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "让手机语音输入无缝连接电脑",
                fontSize = 16.sp,
                color = subTextColor
            )

            Spacer(modifier = Modifier.weight(1f))

            // Guide card
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(20.dp),
                colors = CardDefaults.cardColors(containerColor = cardBg),
                border = androidx.compose.foundation.BorderStroke(1.dp, borderClr)
            ) {
                Column(modifier = Modifier.padding(20.dp)) {
                    Text(
                        text = "输入法安装与配置指南",
                        fontWeight = FontWeight.Bold,
                        fontSize = 16.sp,
                        color = textColor,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )
                    GuideStep("1", "安装推荐输入法", "推荐使用「豆包输入法」或「微信输入法」", textColor, subTextColor)
                    Spacer(modifier = Modifier.height(16.dp))
                    GuideStep("2", "启用键盘", "前往「系统设置」→「语言与输入法」→「管理键盘」", textColor, subTextColor)
                    Spacer(modifier = Modifier.height(16.dp))
                    GuideStep("3", "隐私安全", "Airvoice 仅读取本 App 内的输入框，无需额外权限", textColor, subTextColor)
                }
            }

            Spacer(modifier = Modifier.weight(1f))

            // Start button
            Button(
                onClick = onStartScanning,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                shape = RoundedCornerShape(28.dp),
                colors = ButtonDefaults.buttonColors(containerColor = AppColors.accent)
            ) {
                Text(text = "开始使用", fontSize = 16.sp, fontWeight = FontWeight.Bold, color = Color.White)
            }
        }
    }
}

@Composable
private fun GuideStep(
    number: String,
    title: String,
    desc: String,
    titleColor: Color,
    descColor: Color
) {
    Row(
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        modifier = Modifier.fillMaxWidth()
    ) {
        Box(
            modifier = Modifier
                .size(24.dp)
                .background(AppColors.accent.copy(alpha = 0.2f), CircleShape)
                .border(1.dp, AppColors.accent.copy(alpha = 0.5f), CircleShape),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = number,
                fontSize = 12.sp,
                fontWeight = FontWeight.Bold,
                color = titleColor
            )
        }
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = title,
                fontSize = 14.sp,
                fontWeight = FontWeight.SemiBold,
                color = titleColor
            )
            Text(
                text = desc,
                fontSize = 12.sp,
                color = descColor
            )
        }
    }
}
