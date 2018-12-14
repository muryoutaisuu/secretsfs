// FUSE Input/Output (FIO)
//
// FIO stands for 'FUSE Input/Output' and provides the interface for programming
// FIO plugins for secretsfs. Those FIO plugins need to be registered with the
// fio.RegisterProvider(fm FIOMap) function.
// The FIOMap makes sure that mountpath in the high-top filesystem and FIO plugin
// are always mapped correctly.
// 
// Also initializes variable Log for shared (and consistent) Logging.
package fio

import(
	"github.com/muryoutaisuu/secretsfs/pkg/sfslog"
	"github.com/muryoutaisuu/secretsfs/pkg/store"

	"github.com/spf13/viper"
)

// fiomaps contains all FIOMaps, that map FIOProvider to MountPaths
var fiomaps map[string]*FIOMap = make(map[string]*FIOMap) // oder map[string]FIOMap

// Log contains all the needed Loggers
var Log *sfslog.Log = sfslog.Logger()

// sto contains the currently set store
var sto store.Store

// RegisterProvider registers FIOMaps.
// To be used inside of init() Function of Plugins.
// If FIO is enabled due to configuration, also enable it.
// If not, only register it in Disabled state
func RegisterProvider(fm *FIOMap) {
	fios := viper.GetStringSlice("ENABLED_FIOS")
	for _,f := range fios {
		if f == fm.Provider.FIOPath() {
			fm.Enabled = true
		}
	}
	fiomaps[fm.Provider.FIOPath()] = fm
}

// FIOMaps returns all registered FIO Plugins mapped with their mountpaths.
// Generally used by high-top secretsfs filesystem so it can correctly redirect
// calls to corresponding plugins.
func FIOMaps() map[string]*FIOMap {
	return fiomaps
}





func init() {
	sto = store.GetStore()
}
