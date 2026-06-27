#!/usr/bin/env bash
# Cross-platform JAVA_HOME detection for macOS and Linux
if [ -n "$JAVA_HOME" ] && [ -x "$JAVA_HOME/bin/java" ]; then
  return 0 2>/dev/null || true
fi

if [ -x /usr/libexec/java_home ]; then
  JAVA_HOME="$(/usr/libexec/java_home -F --version 17 2>/dev/null || /usr/libexec/java_home 2>/dev/null)"
fi

# Homebrew fallback (macOS Apple Silicon / Intel)
if [ -z "$JAVA_HOME" ] || [ ! -x "$JAVA_HOME/bin/java" ]; then
  for prefix in /opt/homebrew /usr/local; do
    candidate="$prefix/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home"
    if [ -x "$candidate/bin/java" ]; then
      JAVA_HOME="$candidate"
      break
    fi
  done
fi

# Linux: resolve symlink chain
if [ -z "$JAVA_HOME" ] || [ ! -x "$JAVA_HOME/bin/java" ]; then
  if [ -x /usr/bin/java ]; then
    java_bin="$(readlink -f /usr/bin/java 2>/dev/null || echo /usr/bin/java)"
    candidate="${java_bin%/bin/java}"
    if [ -x "$candidate/bin/java" ]; then
      JAVA_HOME="$candidate"
    fi
  fi
fi

if [ -z "$JAVA_HOME" ] || [ ! -x "$JAVA_HOME/bin/java" ]; then
  echo "ERROR: Could not detect JAVA_HOME. Install JDK 17 first." >&2
  exit 1
fi

export JAVA_HOME
