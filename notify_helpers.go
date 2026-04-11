package blit

import "time"

// Toast convenience methods on App. These wrap App.Send(ToastCmd(...))
// so consumers don't need to construct ToastMsg manually or import time
// for durations.

// ToastInfo sends an informational toast with a 4s duration.
func (a *App) ToastInfo(title, body string) {
	a.Send(ToastCmd(SeverityInfo, title, body, 4*time.Second))
}

// ToastSuccess sends a success toast with a 4s duration.
func (a *App) ToastSuccess(title, body string) {
	a.Send(ToastCmd(SeveritySuccess, title, body, 4*time.Second))
}

// ToastWarn sends a warning toast with a 6s duration.
func (a *App) ToastWarn(title, body string) {
	a.Send(ToastCmd(SeverityWarn, title, body, 6*time.Second))
}

// ToastError sends an error toast with a 6s duration.
func (a *App) ToastError(title, body string) {
	a.Send(ToastCmd(SeverityError, title, body, 6*time.Second))
}
