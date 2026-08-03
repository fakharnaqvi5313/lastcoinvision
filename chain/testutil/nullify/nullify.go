// Package nullify provides methods to init nil values structs for test
// assertion. Test-only (imported solely by *_test.go files, never built
// into lcvd) — its unsafe.Pointer use (flagged by gosec G103) is the
// standard reflect.NewAt pattern for mutating an already-bounds-checked
// struct field/slice element through reflection, not raw pointer
// arithmetic on attacker-influenced data. #nosec G103
package nullify

import (
	"reflect"
	"unsafe"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	coinType  = reflect.TypeOf(sdk.Coin{})
	coinsType = reflect.TypeOf(sdk.Coins{})
)

// Fill analyze all struct fields and slices with
// reflection and initialize the nil and empty slices,
// structs, and pointers.
func Fill(x interface{}) interface{} {
	v := reflect.Indirect(reflect.ValueOf(x))
	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			obj := v.Index(i)
			objPt := reflect.NewAt(obj.Type(), unsafe.Pointer(obj.UnsafeAddr())).Interface() // #nosec G103 -- test-only reflection helper, see package doc
			objPt = Fill(objPt)
			obj.Set(reflect.ValueOf(objPt))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := reflect.Indirect(v.Field(i))
			if !f.CanSet() {
				continue
			}
			switch f.Kind() {
			case reflect.Slice:
				f.Set(reflect.MakeSlice(f.Type(), 0, 0))
			case reflect.Struct:
				switch f.Type() {
				case coinType:
					coin := reflect.New(coinType).Interface()
					s := reflect.ValueOf(coin).Elem()
					f.Set(s)
				case coinsType:
					coins := reflect.New(coinsType).Interface()
					s := reflect.ValueOf(coins).Elem()
					f.Set(s)
				default:
					objPt := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Interface() // #nosec G103 -- test-only reflection helper, see package doc
					s := Fill(objPt)
					f.Set(reflect.ValueOf(s))
				}
			}
		}
	}
	return reflect.Indirect(v).Interface()
}
