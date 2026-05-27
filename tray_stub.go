//go:build !darwin && !windows

package main

// initPlatformTray is a no-op on unsupported platforms.
func (a *App) initPlatformTray() {}

// onBalanceUpdate is a no-op on unsupported platforms.
func (a *App) onBalanceUpdate(_ UsageBalance) {}
