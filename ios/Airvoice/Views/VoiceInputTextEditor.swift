import SwiftUI
import UIKit

final class PreferredKeyboardTextView: UITextView {
    override var textInputMode: UITextInputMode? {
        PreferredKeyboardResolver.preferredInputMode() ?? super.textInputMode
    }

    // No manual editing: voice dictation still inserts text, but the user can't
    // place the caret, select, or invoke the edit menu.
    override func canPerformAction(_ action: Selector, withSender sender: Any?) -> Bool {
        false
    }
}

struct VoiceInputTextEditor: UIViewRepresentable {
    @Binding var text: String
    @Binding var isFocused: Bool
    var textColor: UIColor
    var onTextChange: (String) -> Void

    func makeCoordinator() -> Coordinator {
        Coordinator(parent: self)
    }

    func makeUIView(context: Context) -> PreferredKeyboardTextView {
        let textView = PreferredKeyboardTextView()
        textView.delegate = context.coordinator
        textView.backgroundColor = .clear
        textView.font = .preferredFont(forTextStyle: .body)
        textView.textColor = textColor
        textView.textContainerInset = UIEdgeInsets(top: 8, left: 4, bottom: 8, right: 4)
        textView.autocorrectionType = .yes
        textView.autocapitalizationType = .sentences
        // Keep the keyboard pinned: never allow interactive/scroll dismissal.
        textView.keyboardDismissMode = .none
        // Must stay editable AND interactive: a UITextView can only become first
        // responder (and thus raise the keyboard) when user interaction is enabled.
        // The edit menu is suppressed via `canPerformAction`, so manual editing is
        // still effectively disabled while voice dictation keeps working.
        textView.isEditable = true
        textView.isUserInteractionEnabled = true
        return textView
    }

    func updateUIView(_ textView: PreferredKeyboardTextView, context: Context) {
        if textView.text != text {
            textView.text = text
        }
        if textView.textColor != textColor {
            textView.textColor = textColor
        }

        if isFocused, !textView.isFirstResponder {
            DispatchQueue.main.async {
                textView.becomeFirstResponder()
            }
        } else if !isFocused, textView.isFirstResponder {
            textView.resignFirstResponder()
        }
    }

    final class Coordinator: NSObject, UITextViewDelegate {
        var parent: VoiceInputTextEditor

        init(parent: VoiceInputTextEditor) {
            self.parent = parent
        }

        func textViewDidChange(_ textView: UITextView) {
            let newText = textView.text ?? ""
            parent.text = newText
            parent.onTextChange(newText)
        }

        func textViewDidBeginEditing(_ textView: UITextView) {
            if !parent.isFocused {
                parent.isFocused = true
            }
        }

        func textViewDidEndEditing(_ textView: UITextView) {
            // Stay focused unless focus was intentionally released (e.g. scanner sheet).
            guard parent.isFocused else { return }
            DispatchQueue.main.async {
                guard self.parent.isFocused, !textView.isFirstResponder else { return }
                textView.becomeFirstResponder()
            }
        }
    }
}
