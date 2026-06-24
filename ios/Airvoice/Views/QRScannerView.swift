import SwiftUI
import VisionKit

struct QRScannerView: UIViewControllerRepresentable {
    var onScan: (String) -> Void
    var onError: (Error) -> Void
    
    class Coordinator: NSObject, DataScannerViewControllerDelegate {
        var parent: QRScannerView
        
        init(parent: QRScannerView) {
            self.parent = parent
        }
        
        func dataScanner(_ dataScanner: DataScannerViewController, didAdd addedItems: [RecognizedItem], allItems: [RecognizedItem]) {
            for item in addedItems {
                if case .barcode(let barcode) = item, let stringValue = barcode.payloadStringValue {
                    parent.onScan(stringValue)
                    return
                }
            }
        }
    }
    
    func makeCoordinator() -> Coordinator {
        Coordinator(parent: self)
    }
    
    func makeUIViewController(context: Context) -> DataScannerViewController {
        let scanner = DataScannerViewController(
            recognizedDataTypes: [.barcode(symbologies: [.qr])],
            qualityLevel: .balanced,
            recognizesMultipleItems: false,
            isHighFrameRateTrackingEnabled: false,
            isHighlightingEnabled: true
        )
        scanner.delegate = context.coordinator
        return scanner
    }
    
    func updateUIViewController(_ uiViewController: DataScannerViewController, context: Context) {
        if uiViewController.isSupported && uiViewController.isAvailable {
            do {
                try uiViewController.startScanning()
            } catch {
                onError(error)
            }
        } else {
            onError(NSError(domain: "QRScannerView", code: 1, userInfo: [NSLocalizedDescriptionKey: "VisionKit DataScanner is not supported or available on this device."]))
        }
    }
    
    static func dismantleUIViewController(_ uiViewController: DataScannerViewController, coordinator: Coordinator) {
        uiViewController.stopScanning()
    }
}
