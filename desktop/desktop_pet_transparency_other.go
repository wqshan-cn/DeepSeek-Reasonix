//go:build !windows

package main

const (
	desktopPetTransparencyR uint8 = 0
	desktopPetTransparencyG uint8 = 0
	desktopPetTransparencyB uint8 = 0
)

func enableDesktopPetTransparency() {}
