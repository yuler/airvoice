---
version: alpha
name: Airvoice
description: Airvoice unified design system for mobile apps (GUI) and CLI (Terminal) interfaces.
colors:
  # Light Theme (GUI)
  light:
    primary-text: "#171717"
    secondary-text: "#666666"
    muted-text: "#888888"
    background-primary: "#ffffff"
    background-secondary: "#fafafa"
    border-default: "#eaeaea"
    accent-blue: "#006efe"
    status-success: "#28a948"
    status-warning: "#ffae00"
    status-error: "#fc0035"
    status-neutral: "#8f8f8f"
  # Dark Theme (GUI)
  dark:
    primary-text: "#ededed"
    secondary-text: "#a0a0a0"
    muted-text: "#666666"
    background-primary: "#000000"
    background-secondary: "#0d0e15"
    border-default: "#2e2e2e"
    accent-blue: "#006efe"
    status-success: "#00ac3a"
    status-warning: "#ffae00"
    status-error: "#e2162a"
    status-neutral: "#8f8f8f"
  # Terminal/CLI Mapping
  terminal:
    primary-text: "default"
    secondary-text: "dark-gray"
    muted-text: "dark-gray"
    background-primary: "default"
    background-secondary: "black"
    border-default: "dark-gray"
    accent-blue: "cyan"
    status-success: "green"
    status-warning: "yellow"
    status-error: "red"
    status-neutral: "gray"
typography:
  title:
    mobile:
      fontSize: "24px"
      fontWeight: 600
      letterSpacing: "-0.96px"
    terminal:
      style: "bold/underline"
  body:
    mobile:
      fontSize: "16px"
      fontWeight: 400
    terminal:
      style: "normal"
  detail:
    mobile:
      fontSize: "12px"
      fontWeight: 400
    terminal:
      style: "dim"
  mono:
    mobile:
      fontFamily: "system-mono"
      fontSize: "14px"
    terminal:
      style: "normal"
spacing:
  small:
    mobile: "8px"
    terminal: "1 space"
  medium:
    mobile: "16px"
    terminal: "1 empty line / 2 spaces"
  large:
    mobile: "24px"
    terminal: "2 empty lines / 4 spaces"
rounded:
  sm: "6px"
  md: "12px"
  lg: "16px"
  full: "9999px"
components:
  button-primary:
    backgroundColor: "{colors.accent-blue}"
    textColor: "#ffffff"
    rounded: "{rounded.full}"
    height: "56px"
  button-secondary:
    backgroundColor: "transparent"
    textColor: "{colors.primary-text}"
    rounded: "{rounded.full}"
    borderColor: "{colors.border-default}"
    height: "44px"
  badge:
    rounded: "{rounded.md}"
    padding: "6px 12px"
  input:
    backgroundColor: "{colors.background-secondary}"
    rounded: "{rounded.lg}"
    borderColor: "{colors.border-default}"
---

# Airvoice Design System

## Overview

Airvoice is a minimalist utility bridge. Its design system emphasizes clarity, high contrast, and focus. The aesthetic is strictly monochrome (pure blacks, grays, and whites) accented by a single brand color (**Vercel Blue**) to denote connectivity and primary actions.

This design system is unified: every token maps to both **Graphical User Interfaces (iOS / Android)** and the **Command Line Interface (Terminal)**.

---

## Colors & Themes

Each non-neutral status color is reserved exclusively for state signaling:

- **Accent Blue (`#006efe`)**: Indicates active pairing, successful connection focus, and primary button actions.
- **Success (`#00ac3a` / `#28a948`)**: Device connected.
- **Warning (`#ffae00`)**: Connecting in progress.
- **Error (`#e2162a` / `#fc0035`)**: Connection/pairing failed or permission required.
- **Neutral/Muted**: Inactive, offline, or disconnected states.

### GUI Surfaces
In Dark Mode, backgrounds must be pure black (`#000000`). Textareas and containers use a subtle secondary layer (`#0d0e15`) with a 1px border (`#2e2e2e`) instead of drop shadows to maintain a clean, flat aesthetic.

### CLI Mapping
Terminal colors use standard ANSI codes mapping to their semantic equivalent. The background remains default, utilizing terminal transparency/defaults where possible, with gray dividers.

---

## Typography & Spacing

### Mobile GUI
- **Title**: Large, bold, tracking-tight headline text for onboarding pages and pairing statuses.
- **Body**: Default text for text editor input and messages.
- **Detail**: Captions, auxiliary text below buttons.
- **Mono**: IP addresses, WebSocket URIs, and pairing tokens.

### Terminal CLI
- **Title**: Bold/Underlined section headers.
- **Detail**: Dimmed gray output for logs and secondary instructions.

### Spacing Grid
Spacing enforces a strict three-level rhythm:
- **Small (8px / 1 char)**: Inline elements (icon to text, status dot to text).
- **Medium (16px / 1 empty line)**: Block elements (badge to title, input area to buttons).
- **Large (24px / 2 empty lines)**: Section-level gaps (header block to input area).

---

## Components & UI States

### Buttons
- **Primary (Pill)**: Filled with `accent-blue` and white text. Uses `rounded.full` (capsule). Highly visible.
- **Secondary (Bordered)**: Transparent fill, bordered by `border-default`, carrying secondary actions (e.g., "Manual Send").

### Inputs
- Text editor areas use `rounded.lg` (16px rounded corners) with a subtle `background-secondary` background and `border-default` stroke.

### State Matrix

| State | GUI Presentation | Terminal Presentation |
| :--- | :--- | :--- |
| **Default** | Primary text, default borders, flat background | Default fore/background text |
| **Hover** | 10% opacity white overlay or brighter border | Text is underlined or highlighted |
| **Focus** | 2px background-gap + 2px `accent-blue` outer ring | Left-padded with `>` character |
| **Disabled**| 30% opacity, standard default cursor, grayed out | 30% opacity/dimmed, unselectable |

---

## Command Line Interface (CLI) Specifics

1. **Margins & Padding**: Indent all terminal output text by `1 space` (`spacing.small`) from the left terminal edge. Avoid print statements directly touching the terminal border.
2. **QR Code Generation**: 
   - Print terminal pairing QR codes with a minimum border width of `2 spaces` on all sides (Quiet Zone).
   - Use the terminal default foreground and background settings rather than inverted colors to prevent barcode scanners from failing on dark terminal windows.
3. **Interactive Menus**: Use standard arrow indicators (`>`) with the select line highlighted in `accent-blue` (Cyan).

---

## Voice & Tone

- **Concise & Verb-First**: All button actions must start with a verb indicating the direct action (e.g., `Pair Device`, not `OK` or `Submit`).
- **No Marketing Filler**: Do not use words like "Please" or "Successfully".
  - *Correct*: `Device paired`
  - *Incorrect*: `Device has been successfully paired. Please check your desktop.`
- **Progress States**: Display active operations using the present participle with three dots: `Sending...`, `Connecting...`.

---

## Do's & Don'ts

### Do
- Always show a visible connection badge (`badge` style) indicating status (`Success`/`Warning`/`Error`/`Neutral`) at the top of the main screen.
- Keep layout borders sharp and clean.
- Ensure the active input area takes focus ring indicators during text editing.

### Don't
- Do not mix rounded and sharp corners. Stick to the designated `rounded` tokens.
- Do not use gradients or colored backgrounds for text boxes. Keep surfaces black/dark gray.
- Do not use status colors for purely decorative purposes. Color must carry functional meaning.
