package fio

import(
	"github.com/Muryoutaisuu/secretsfs/pkg/sfslog"
	"github.com/Muryoutaisuu/secretsfs/pkg/store"
)

// fiomaps contains all FIOMaps, that map FIOProvider to MountPaths
//var fiomaps []*FIOMap // oder map[string]FIOMap
var fiomaps map[string]*FIOMap = make(map[string]*FIOMap) // oder map[string]FIOMap

// Log contains all the needed Loggers
var Log *sfslog.Log = sfslog.Logger()

// sto contains the currently set store
var sto store.Store

// RegisterProvider registers FIOMaps
func RegisterProvider(fm *FIOMap) {
	fiomaps[fm.MountPath] = fm
	//fiomaps = append(fiomaps, fm)
}

func FIOMaps() map[string]*FIOMap {
	return fiomaps
}





func init() {
	sto = store.GetStore()
}
