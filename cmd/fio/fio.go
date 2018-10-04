package fio

// fiomaps contains all FIOMaps, that map FIOProvider to MountPaths
var fiomaps []FIOMap // oder map[string]FIOMap

// RegisterProvder registers FIOMaps
func RegisterProvider(fm *FIOMap) {
       fiomaps = append(fiomaps, fm)
}

