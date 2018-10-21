package fio

import(
	"github.com/Muryoutaisuu/secretsfs/pkg/sfslog"
)

// fiomaps contains all FIOMaps, that map FIOProvider to MountPaths
//var fiomaps []*FIOMap // oder map[string]FIOMap
var fiomaps map[string]*FIOMap = make(map[string]*FIOMap) // oder map[string]FIOMap

// Log contains all the needed Loggers
var Log *sfslog.Log = sfslog.Logger()

// RegisterProvider registers FIOMaps
func RegisterProvider(fm *FIOMap) {
	fiomaps[fm.MountPath] = fm
	//fiomaps = append(fiomaps, fm)
}

func FIOMaps() map[string]*FIOMap {
	return fiomaps
}
