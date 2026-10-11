package cbgo_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tinygo-org/cbgo"
	"github.com/tinygo-org/cbgo/internal/heap"
)

// TestEveryCallRunsInAPool checks that every function calling into
// Objective-C starts with `defer popPool(pushPool())`. Without it, what the
// call autoreleases stays allocated for as long as its thread lives, and Go
// threads rarely exit. The tests below check representative calls on the C
// heap; this one keeps the rest, and any added later, from going without.
func TestEveryCallRunsInAPool(t *testing.T) {
	names, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || fn.Name.Name == "pushPool" || fn.Name.Name == "popPool" {
				continue
			}
			if callsObjC(fn.Body) && !startsWithPool(fn.Body) {
				t.Errorf("%s: %s calls into Objective-C outside an autorelease pool", fset.Position(fn.Pos()), fn.Name.Name)
			}
		}
	}
}

// callsObjC reports whether body calls one of the package's C functions, all
// of which are named cb_ or bt_.
func callsObjC(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "C" &&
					(strings.HasPrefix(sel.Sel.Name, "cb_") || strings.HasPrefix(sel.Sel.Name, "bt_")) {
					found = true
				}
			}
		}
		return !found
	})
	return found
}

func startsWithPool(body *ast.BlockStmt) bool {
	if len(body.List) == 0 {
		return false
	}
	d, ok := body.List[0].(*ast.DeferStmt)
	return ok && types.ExprString(d.Call) == "popPool(pushPool())"
}

// A test repeats a call in batches to see what it leaves behind.
const (
	batches = 5
	calls   = 2000 // per batch
)

// releases fails the test if calling fn leaves objects on the C heap. An
// object left in an autorelease pool costs at least its 8-byte pool slot, so
// a call that leaks one leaves 8 bytes or more each time, in every batch. The
// allocator's own count also moves, now and then and in 16 KiB steps on
// arm64, but not in every batch: a test fails only when every batch left more
// than 1 byte per call.
func releases(t *testing.T, fn func()) {
	t.Helper()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	fn() // The first call fills caches that every later call shares.
	least := -1
	for b := 0; b < batches; b++ {
		before := heap.InUse()
		for i := 0; i < calls; i++ {
			fn()
		}
		if left := heap.InUse() - before; least < 0 || left < least {
			least = left
		}
	}
	if least > calls {
		t.Errorf("each batch of %d calls left at least %d bytes on the C heap (%d per call)", calls, least, least/calls)
	}
}

// Messages to a nil object do nothing in Objective-C, so the zero values
// below run each wrapper's own work without a Bluetooth controller.

func TestWriteCharacteristicReleasesValue(t *testing.T) {
	value := make([]byte, 244)
	releases(t, func() {
		cbgo.Peripheral{}.WriteCharacteristic(value, cbgo.Characteristic{}, false)
	})
}

// A 16-bit UUID is short enough to become a tagged-pointer NSString, which is
// never allocated, so the tests use 128-bit ones too.
var (
	uuid16  = cbgo.UUID16(0xfe40)
	uuid128 = cbgo.MustParseUUID("00467768-6228-2272-4663-277478268000")
)

func TestScanReleasesFilters(t *testing.T) {
	releases(t, func() {
		cbgo.CentralManager{}.Scan([]cbgo.UUID{uuid16, uuid128}, &cbgo.CentralManagerScanOpts{
			SolicitedServiceUUIDs: []cbgo.UUID{uuid16, uuid128},
		})
	})
}

func TestRetrievePeripheralsWithIdentifiersReleasesIdentifiers(t *testing.T) {
	releases(t, func() {
		cbgo.CentralManager{}.RetrievePeripheralsWithIdentifiers([]cbgo.UUID{uuid128})
	})
}

func TestRetrieveConnectedPeripheralsWithServicesReleasesServices(t *testing.T) {
	releases(t, func() {
		cbgo.CentralManager{}.RetrieveConnectedPeripheralsWithServices([]cbgo.UUID{uuid16, uuid128})
	})
}

func TestCharacteristicUUIDReleasesString(t *testing.T) {
	chr := cbgo.NewMutableCharacteristic(uuid128, cbgo.CharacteristicPropertyRead, nil, cbgo.AttributePermissionsReadable).Characteristic()
	releases(t, func() {
		if got := chr.UUID(); got.String() != uuid128.String() {
			t.Fatalf("UUID() = %s, want %s", got, uuid128)
		}
	})
}

func TestServiceUUIDReleasesString(t *testing.T) {
	svc := cbgo.NewMutableService(uuid128, true).Service()
	releases(t, func() {
		if got := svc.UUID(); got.String() != uuid128.String() {
			t.Fatalf("UUID() = %s, want %s", got, uuid128)
		}
	})
}

func TestStartAdvertisingReleasesData(t *testing.T) {
	ad := cbgo.AdvData{LocalName: "A long local name", ServiceUUIDs: []cbgo.UUID{uuid16, uuid128}}
	releases(t, func() {
		cbgo.PeripheralManager{}.StartAdvertising(ad)
	})
}

func TestStartAdvertisingReleasesBeacon(t *testing.T) {
	ad := cbgo.AdvData{IBeaconData: make([]byte, 21)}
	releases(t, func() {
		cbgo.PeripheralManager{}.StartAdvertising(ad)
	})
}

func TestUpdateValueReleasesValue(t *testing.T) {
	value := make([]byte, 244)
	releases(t, func() {
		cbgo.PeripheralManager{}.UpdateValue(value, cbgo.Characteristic{}, nil)
	})
}

func TestSetCharacteristicsReleasesArray(t *testing.T) {
	svc := cbgo.NewMutableService(uuid128, true)
	chr := cbgo.NewMutableCharacteristic(uuid128, cbgo.CharacteristicPropertyRead, nil, cbgo.AttributePermissionsReadable)
	releases(t, func() {
		svc.SetCharacteristics([]cbgo.MutableCharacteristic{chr})
	})
	if n := len(svc.Service().Characteristics()); n != 1 {
		t.Errorf("the service has %d characteristics, want 1", n)
	}
}

func TestSetIncludedServicesReleasesArray(t *testing.T) {
	svc := cbgo.NewMutableService(uuid128, true)
	inc := cbgo.NewMutableService(uuid16, false)
	releases(t, func() {
		svc.SetIncludedServices([]cbgo.MutableService{inc})
	})
	if n := len(svc.Service().IncludedServices()); n != 1 {
		t.Errorf("the service includes %d services, want 1", n)
	}
}

func TestSetDescriptorsReleasesArray(t *testing.T) {
	chr := cbgo.NewMutableCharacteristic(uuid128, cbgo.CharacteristicPropertyRead, nil, cbgo.AttributePermissionsReadable)
	format := cbgo.NewMutableDescriptor(cbgo.UUID16(0x2904), make([]byte, 7)) // Presentation Format takes bytes
	releases(t, func() {
		chr.SetDescriptors([]cbgo.MutableDescriptor{format})
	})
	if n := len(chr.Characteristic().Descriptors()); n != 1 {
		t.Errorf("the characteristic has %d descriptors, want 1", n)
	}
}
