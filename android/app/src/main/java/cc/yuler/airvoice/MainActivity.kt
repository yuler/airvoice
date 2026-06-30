package cc.yuler.airvoice

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.viewModels
import cc.yuler.airvoice.ui.screens.MainScreen
import cc.yuler.airvoice.ui.viewmodel.AirvoiceViewModel

class MainActivity : ComponentActivity() {
    private val viewModel: AirvoiceViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            MainScreen(viewModel = viewModel)
        }
    }
}
