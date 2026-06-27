package com.yule.airvoice.ui.screens

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
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
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = "Airvoice",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = textColor,
                modifier = Modifier.padding(bottom = 8.dp)
            )
            Text(
                text = "让手机语音输入无缝连接电脑",
                fontSize = 16.sp,
                color = subTextColor,
                modifier = Modifier.padding(bottom = 32.dp)
            )

            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 48.dp),
                shape = RoundedCornerShape(20.dp),
                colors = CardDefaults.cardColors(containerColor = cardBg),
                border = androidx.compose.foundation.BorderStroke(1.dp, borderClr)
            ) {
                Column(modifier = Modifier.padding(20.dp)) {
                    Text(
                        text = "连接步骤:",
                        fontWeight = FontWeight.Bold,
                        color = textColor,
                        modifier = Modifier.padding(bottom = 12.dp)
                    )
                    GuideStep("1", "在电脑上打开终端，运行 `airvoice dev` 启动服务", textColor, subTextColor)
                    Spacer(modifier = Modifier.height(12.dp))
                    GuideStep("2", "点击下方按钮，扫描电脑终端里生成的二维码", textColor, subTextColor)
                    Spacer(modifier = Modifier.height(12.dp))
                    GuideStep("3", "使用手机键盘的语音输入法说话，电脑即可实时打字", textColor, subTextColor)
                }
            }

            Button(
                onClick = onStartScanning,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                shape = RoundedCornerShape(28.dp),
                colors = ButtonDefaults.buttonColors(containerColor = AppColors.accent)
            ) {
                Text(text = "扫码连接电脑", fontSize = 16.sp, color = Color.White)
            }
        }
    }
}

@Composable
private fun GuideStep(number: String, text: String, titleColor: Color, descColor: Color) {
    Row(
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        modifier = Modifier.fillMaxWidth()
    ) {
        Box(
            modifier = Modifier
                .size(24.dp)
                .background(AppColors.accent.copy(alpha = 0.2f), shape = androidx.compose.foundation.shape.CircleShape),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = number,
                fontSize = 12.sp,
                fontWeight = FontWeight.Bold,
                color = titleColor
            )
        }
        Text(
            text = text,
            fontSize = 14.sp,
            color = descColor,
            modifier = Modifier.weight(1f)
        )
    }
}
