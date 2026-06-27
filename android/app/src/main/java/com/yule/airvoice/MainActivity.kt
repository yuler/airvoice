package com.yule.airvoice

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.viewModels
import com.yule.airvoice.ui.screens.MainScreen
import com.yule.airvoice.ui.viewmodel.AirvoiceViewModel

class MainActivity : ComponentActivity() {
    private val viewModel: AirvoiceViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            MainScreen(viewModel = viewModel)
        }
    }
}
