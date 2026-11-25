import SwiftUI
import UIKit

struct SettingsSheetView: View {
    @ObservedObject var settings: AppSettings
    @ObservedObject var connection: ConnectionManager
    let theme: AppTheme
    let onDisconnectAndClear: () -> Void
    let onSettingsReset: () -> Void

    @Environment(\.dismiss) private var dismiss

    @State private var showDisconnectConfirm = false
    @State private var showResetConfirm = false
    @State private var toastMessage: String?
    @State private var isToastError = false

    private var strings: AppStrings { settings.strings }

    private var appVersion: String {
        Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "0.0.0"
    }

    private var systemVersion: String {
        "\(UIDevice.current.systemName) \(UIDevice.current.systemVersion)"
    }

    var body: some View {
        NavigationStack {
            ZStack {
                theme.background.ignoresSafeArea()

                ScrollView {
                    VStack(spacing: 20) {
                        appearanceSection
                        generalSection
                        connectionSection
                        aboutSection
                    }
                    .padding(.horizontal, 20)
                    .padding(.top, 12)
                    .padding(.bottom, 28)
                }
            }
            .navigationTitle(strings.settingsTitle)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button(strings.done) {
                        dismiss()
                    }
                    .fontWeight(.semibold)
                    .foregroundStyle(theme.accent)
                }
            }
            .toast(message: $toastMessage, isError: $isToastError, theme: theme)
            .alert(strings.disconnectConfirmTitle, isPresented: $showDisconnectConfirm) {
                Button(strings.cancel, role: .cancel) {}
                Button(strings.disconnectConfirmAction, role: .destructive) {
                    onDisconnectAndClear()
                    showLocalToast(strings.toastPairingCleared)
                }
            } message: {
                Text(strings.disconnectConfirmBody)
            }
            .alert(strings.resetConfirmTitle, isPresented: $showResetConfirm) {
                Button(strings.cancel, role: .cancel) {}
                Button(strings.resetConfirmAction) {
                    settings.resetToDefaults()
                    onSettingsReset()
                    showLocalToast(settings.strings.toastSettingsReset)
                }
            } message: {
                Text(strings.resetConfirmBody)
            }
        }
        .preferredColorScheme(settings.themeMode.preferredColorScheme)
        .presentationDetents([.large])
        .presentationDragIndicator(.visible)
    }

    // MARK: - Sections

    private var appearanceSection: some View {
        SettingsSectionCard(title: strings.sectionAppearance, theme: theme) {
            SettingsLabelledControl(label: strings.themeLabel, theme: theme) {
                SettingsSegmentedControl(
                    options: [
                        (ThemeMode.system, strings.themeSystem),
                        (ThemeMode.light, strings.themeLight),
                        (ThemeMode.dark, strings.themeDark),
                    ],
                    selected: settings.themeMode,
                    theme: theme
                ) { mode in
                    settings.themeMode = mode
                }
            }

            sectionDivider

            SettingsLabelledControl(label: strings.languageLabel, theme: theme) {
                SettingsSegmentedControl(
                    options: [
                        (AppLanguage.system, strings.languageSystem),
                        (AppLanguage.zh, strings.languageZh),
                        (AppLanguage.en, strings.languageEn),
                    ],
                    selected: settings.language,
                    theme: theme
                ) { language in
                    settings.language = language
                }
            }
        }
    }

    private var generalSection: some View {
        SettingsSectionCard(title: strings.sectionGeneral, theme: theme) {
            toggleRow(title: strings.autoSendLabel, isOn: $settings.autoSendEnabled)

            if settings.autoSendEnabled {
                sectionDivider
                SettingsLabelledControl(label: strings.autoSendDelayLabel, theme: theme) {
                    SettingsSegmentedControl(
                        options: AppSettings.autoSendDelayOptions.map { ms in
                            (ms, formatSeconds(ms))
                        },
                        selected: settings.autoSendDelayMs,
                        theme: theme
                    ) { delayMs in
                        settings.autoSendDelayMs = delayMs
                    }
                }
            }

            sectionDivider
            toggleRow(title: strings.hapticFeedbackLabel, isOn: $settings.hapticFeedbackEnabled)

            sectionDivider
            toggleRow(title: strings.keepScreenAwakeLabel, isOn: $settings.keepScreenAwake)

            sectionDivider
            toggleRow(title: strings.showInputTipsLabel, isOn: $settings.showInputTips)

            sectionDivider
            SettingsLabelledControl(label: strings.sendTimeoutLabel, theme: theme) {
                SettingsSegmentedControl(
                    options: AppSettings.sendTimeoutOptions.map { ms in
                        (ms, "\(Int(ms / 1000))s")
                    },
                    selected: settings.sendTimeoutMs,
                    theme: theme
                ) { timeoutMs in
                    settings.sendTimeoutMs = timeoutMs
                }
            }
        }
    }

    private var connectionSection: some View {
        SettingsSectionCard(title: strings.sectionConnection, theme: theme) {
            VStack(alignment: .leading, spacing: 4) {
                Text(strings.pairedServerLabel)
                    .font(.subheadline)
                    .foregroundStyle(theme.primaryText)

                Text(connection.savedServerAddress ?? strings.notPaired)
                    .font(connection.savedServerAddress != nil ? .system(size: 13, design: .monospaced) : .footnote)
                    .foregroundStyle(theme.secondaryText)
                    .lineLimit(1)
                    .truncationMode(.middle)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal, 14)
            .padding(.vertical, 12)

            sectionDivider

            let hasPairing = connection.savedServerAddress != nil
            Button {
                showDisconnectConfirm = true
            } label: {
                Text(strings.disconnectAndClear)
                    .font(.subheadline.weight(.medium))
                    .foregroundStyle(theme.statusBarError.opacity(hasPairing ? 1.0 : 0.30))
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 13)
            }
            .disabled(!hasPairing)

            sectionDivider

            Button {
                showResetConfirm = true
            } label: {
                Text(strings.resetDefaults)
                    .font(.subheadline.weight(.medium))
                    .foregroundStyle(theme.primaryText)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 13)
            }
        }
    }

    private var aboutSection: some View {
        SettingsSectionCard(title: strings.sectionAbout, theme: theme) {
            Button {
                let diagnostic = "Airvoice v\(appVersion) · \(systemVersion)"
                UIPasteboard.general.string = diagnostic
                showLocalToast(strings.toastVersionCopied)
            } label: {
                VStack(spacing: 8) {
                    HStack {
                        Text(strings.appVersionLabel)
                            .font(.subheadline)
                            .foregroundStyle(theme.primaryText)
                        Spacer()
                        Text("v\(appVersion)")
                            .font(.system(size: 13, design: .monospaced))
                            .foregroundStyle(theme.secondaryText)
                    }

                    HStack {
                        Text(strings.systemVersionLabel)
                            .font(.subheadline)
                            .foregroundStyle(theme.primaryText)
                        Spacer()
                        Text(systemVersion)
                            .font(.system(size: 13, design: .monospaced))
                            .foregroundStyle(theme.secondaryText)
                    }

                    HStack(spacing: 4) {
                        Spacer()
                        Image(systemName: "doc.on.doc")
                            .font(.system(size: 11))
                        Text(strings.tapToCopyHint)
                            .font(.caption2)
                    }
                    .foregroundStyle(theme.secondaryText)
                }
                .padding(.horizontal, 14)
                .padding(.vertical, 12)
            }
            .buttonStyle(.plain)
        }
    }

    // MARK: - Helpers

    private func toggleRow(title: String, isOn: Binding<Bool>) -> some View {
        Toggle(isOn: isOn) {
            Text(title)
                .font(.subheadline)
                .foregroundStyle(theme.primaryText)
        }
        .tint(theme.accent)
        .padding(.horizontal, 14)
        .padding(.vertical, 10)
    }

    private var sectionDivider: some View {
        Rectangle()
            .fill(theme.border)
            .frame(height: 1)
            .padding(.horizontal, 14)
    }

    private func showLocalToast(_ message: String) {
        isToastError = false
        toastMessage = message
    }

    private func formatSeconds(_ ms: Int) -> String {
        String(format: "%.1fs", Double(ms) / 1000.0)
    }
}

private struct SettingsSectionCard<Content: View>: View {
    let title: String
    let theme: AppTheme
    @ViewBuilder let content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.caption.weight(.medium))
                .foregroundStyle(theme.secondaryText)
                .padding(.horizontal, 4)

            VStack(spacing: 0) {
                content()
            }
            .background(theme.secondaryBackground)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(theme.border, lineWidth: 1)
            )
        }
    }
}

private struct SettingsLabelledControl<Content: View>: View {
    let label: String
    let theme: AppTheme
    @ViewBuilder let content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(label)
                .font(.subheadline)
                .foregroundStyle(theme.primaryText)

            content()
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 12)
    }
}

private struct SettingsSegmentedControl<T: Equatable>: View {
    let options: [(T, String)]
    let selected: T
    let theme: AppTheme
    let onSelect: (T) -> Void

    var body: some View {
        HStack(spacing: 4) {
            ForEach(Array(options.enumerated()), id: \.offset) { _, item in
                let isSelected = item.0 == selected
                Button {
                    onSelect(item.0)
                } label: {
                    Text(item.1)
                        .font(.footnote.weight(isSelected ? .semibold : .regular))
                        .foregroundStyle(isSelected ? Color.white : theme.primaryText)
                        .lineLimit(1)
                        .minimumScaleFactor(0.85)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 7)
                        .background(isSelected ? theme.accent : Color.clear)
                        .clipShape(RoundedRectangle(cornerRadius: 6))
                }
                .buttonStyle(.plain)
            }
        }
        .padding(3)
        .background(theme.chipBackground)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}
