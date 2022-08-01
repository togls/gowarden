package model

import "time"

type TwoFactorIncomplete struct {
	UserUuid   string
	DeviceUuid string
	DeviceName string
	LoginTime  time.Time
	IpAddress  string
}
