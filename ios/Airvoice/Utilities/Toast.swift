import SwiftUI

struct Toast: View {
    let message: String
    let isError: Bool
    
    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: isError ? "exclamationmark.triangle.fill" : "checkmark.circle.fill")
                .foregroundStyle(isError ? .red : .green)
            
            Text(message)
                .font(.subheadline)
                .foregroundColor(.white)
                .lineLimit(2)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(
            Capsule()
                .fill(Color(hex: "1F2030").opacity(0.95))
                .overlay(
                    Capsule()
                        .stroke(Color.white.opacity(0.1), lineWidth: 1)
                )
        )
        .shadow(color: Color.black.opacity(0.3), radius: 10, y: 5)
    }
}

struct ToastModifier: ViewModifier {
    @Binding var toastMessage: String?
    @Binding var isError: Bool
    
    @State private var dismissTask: Task<Void, Never>? = nil
    
    func body(content: Content) -> some View {
        ZStack {
            content
            
            VStack {
                Spacer()
                if let message = toastMessage {
                    Toast(message: message, isError: isError)
                        .transition(.move(edge: .bottom).combined(with: .opacity))
                        .padding(.bottom, 50)
                }
            }
            .animation(.spring(), value: toastMessage)
        }
        .onChange(of: toastMessage) { oldValue, newValue in
            dismissTask?.cancel()
            if newValue != nil {
                dismissTask = Task {
                    try? await Task.sleep(nanoseconds: 2_000_000_000)
                    guard !Task.isCancelled else { return }
                    withAnimation {
                        toastMessage = nil
                    }
                }
            }
        }
    }
}

extension View {
    func toast(message: Binding<String?>, isError: Binding<Bool>) -> some View {
        self.modifier(ToastModifier(toastMessage: message, isError: isError))
    }
}
